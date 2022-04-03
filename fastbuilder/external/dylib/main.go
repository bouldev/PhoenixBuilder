package main

import (
	"C"
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"phoenixbuilder/fastbuilder/external/connection"
	"phoenixbuilder/fastbuilder/external/packet"
	"phoenixbuilder/minecraft/protocol"
	mc_packet "phoenixbuilder/minecraft/protocol/packet"
	"time"
)

var ErrSendOnClosedConnection = fmt.Errorf("send on closed connection")
var ErrRecvOnClosedConnection = fmt.Errorf("recv on closed connection")

type Client struct {
	conn         connection.ReliableConnection
	close        chan struct{}
	byeSent      bool
	closed       bool
	pongDeadline time.Time
	gamePackets  chan []byte
	mcTypePool   mc_packet.Pool
}

func (c *Client) SendFrame(f []byte) error {
	if c.closed {
		return ErrSendOnClosedConnection
	}

	complete := make(chan struct{})
	go func() {
		err := c.conn.SendFrame(f)
		if err != nil {
			c.Close()
		}
		close(complete)
	}()
	select {
	case <-complete:
		if c.closed {
			fmt.Println("Connection Closed")
			return fmt.Errorf("connection Close")
		}
		return nil
	case <-time.After(1 * time.Second):
		fmt.Println("Send Timeout")
		c.Close()
		return fmt.Errorf("send Timeout")
	}
}

func (c *Client) SendFBCmd(cmd string) error {
	return c.SendFrame(packet.Serialize(&packet.EvalPBCommandPacket{Command: cmd}))
}

func (c *Client) SendMCPacket(pk mc_packet.Packet) error {
	b := &bytes.Buffer{}
	w := protocol.NewWriter(b, 0)
	hdr := pk.ID()
	w.Varuint32(&hdr)
	pk.Marshal(w)
	p := &packet.GamePacket{Content: b.Bytes()}
	return c.Send(p)
}

func (c *Client) sendMCCmd(CommandType byte, cmd string) (uuid.UUID, error) {
	uid, _ := uuid.NewUUID()
	err := c.Send(&packet.GameCommandPacket{
		UUID:        uid,
		CommandType: CommandType,
		Command:     cmd,
	})
	return uid, err
}

func (c *Client) SendNoResponseMCCmd(cmd string) error {
	_, err := c.sendMCCmd(packet.CommandTypeSettings, cmd)
	return err
}

func (c *Client) SendWSCmd(cmd string) (uuid.UUID, error) {
	return c.sendMCCmd(packet.CommandTypeWebsocket, cmd)
}
func (c *Client) SendMCCmd(cmd string) (uuid.UUID, error) {
	return c.sendMCCmd(packet.CommandTypeNormal, cmd)
}

func (c *Client) Send(pk packet.Packet) error {
	return c.SendFrame(packet.Serialize(pk))
}

func (c *Client) RecvGamePacket() ([]byte, error) {
	if c.closed {
		return nil, ErrRecvOnClosedConnection
	}
	select {
	case pk := <-c.gamePackets:
		return pk, nil
	case <-c.close:
		return nil, ErrRecvOnClosedConnection
	}
}

func (c *Client) RecvDecodedGamePacket() (mc_packet.Packet, error) {
	mcPkt, err := c.RecvGamePacket()
	if err != nil {
		return nil, err
	}
	pk := c.mcTypePool[uint32(mcPkt[0])]()
	pk.Unmarshal(protocol.NewReader(bytes.NewReader(mcPkt[1:]), 0))
	return pk, nil
}

func (c *Client) Close() {
	if !c.closed {
		if !c.byeSent {
			go func() {
				packet.SerializeAndSend(&packet.ByePacket{}, c.conn)
			}()
			c.byeSent = true
		}
		close(c.close)
		c.closed = true
		fmt.Println("Closed")
	}
}

func (c *Client) routine() {
	t := time.NewTicker(time.Second)
	rc := make(chan []byte, 1024)
	go func() {
		for {
			select {
			case <-c.close:
				c.Close()
				break
			default:
				frame, err := c.conn.RecvFrame()
				if err != nil {
					c.Close()
					return
				}
				rc <- frame
			}
		}
	}()
	for {
		select {
		case <-t.C:
			if time.Now().After(c.pongDeadline) {
				fmt.Println("Deadline Exceed!")
				c.Close()
				break
			}
			c.Send(&packet.PingPacket{})
			//fmt.Println("Ping")
		case rawPacket := <-rc:
			pkt, canParse := packet.Deserialize(rawPacket)
			if !canParse {
				fmt.Println("Unparsable packet received!")
				c.Close()
				break
			}
			switch p := pkt.(type) {
			case *packet.PongPacket:
				c.pongDeadline = time.Now().Add(3 * time.Second)
				//fmt.Println("Pong")
			case *packet.ByePacket:
				fmt.Println("Connection Terminated By FB")
				c.Close()
			case *packet.GamePacket:
				c.gamePackets <- p.Content
			}
		}
	}
}

func NewClient(address string) *Client {
	retChan := make(chan connection.ReliableConnection)
	go func() {
		conn, err := connection.KCPDial(address)
		if err != nil {
			retChan <- nil
		}
		retChan <- conn
	}()
	select {
	case <-time.After(3 * time.Second):
		return nil
	case conn := <-retChan:
		if conn == nil {
			return nil
		}
		c := &Client{
			conn:         conn,
			close:        make(chan struct{}),
			closed:       false,
			pongDeadline: time.Now().Add(3 * time.Second),
			gamePackets:  make(chan []byte, 1024),
			mcTypePool:   mc_packet.NewPool(),
		}
		go c.routine()
		return c
	}
}

var DontGCMe []*Client

func getConnID(client *Client) int {
	for i, c := range DontGCMe {
		if c == nil {
			DontGCMe[i] = client
			return i
		}
	}
	i := len(DontGCMe)
	DontGCMe = append(DontGCMe, client)
	return i
}

//export ConnectFB
func ConnectFB(address *C.char) int {
	str := C.GoString(address)
	// fmt.Println(str)
	client := NewClient(str)
	if client != nil {
		fmt.Printf("Connect to %v fail\n", str)
		return -1
	}
	return getConnID(client)
}

//export ReleaseConnByID
func ReleaseConnByID(id int) {
	if id >= len(DontGCMe) {
		fmt.Printf("id %v out of range %v\n", id, len(DontGCMe))
		return
	}
	conn := DontGCMe[id]
	if conn != nil {
		conn.Close()
		DontGCMe[id] = nil
	}
	for {
		if len(DontGCMe) != 0 && DontGCMe[len(DontGCMe)-1] == nil {
			DontGCMe = DontGCMe[0 : len(DontGCMe)-1]
		}
	}
}

//
////export RecvFrame
//func RecvFrame(connID int) *C.char {
//	bs, err := DontGCMe[connID].RecvFrame()
//	// fmt.Println(bs, err)
//	if err != nil {
//		// fmt.Println(err)
//		bs = []byte{}
//		ReleaseConnByID(connID)
//	} else {
//		bs = bs[1:]
//	}
//	return C.CString(string(bs))
//}

func main() {
	//go build -o fb_conn.so -buildmode=c-shared main.go
	client := NewClient("124.222.6.29:3456")
	if client != nil {
		fmt.Println("Connect Success")
	} else {
		panic("Connection Fail")
	}
	//client.SendFBCmd("set 0 0 0")
	client.SendMCPacket(&mc_packet.SettingsCommand{
		CommandLine:    "time set night",
		SuppressOutput: true,
	})
	//client.SendWSCmd("list")
	for {
		gamePacket, err := client.RecvDecodedGamePacket()
		if err != nil {
			panic(err)
		}
		fmt.Println(gamePacket)
	}

	//oldP := &packet.EvalPBCommandPacket{Command: "set 0 0 0"}
	//newP := &packet.EvalPBCommandPacket{}
	//newP.Parse(oldP.Marshal())
	//println(newP)
}
