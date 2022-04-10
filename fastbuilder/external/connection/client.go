package connection

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"io"
	"phoenixbuilder/fastbuilder/external/packet"
	"phoenixbuilder/fastbuilder/uqHolder"
	"phoenixbuilder/minecraft/protocol"
	mc_packet "phoenixbuilder/minecraft/protocol/packet"
	"time"
)

var ErrSendOnClosedConnection = fmt.Errorf("send on closed connection")
var ErrRecvOnClosedConnection = fmt.Errorf("recv on closed connection")
var TypePool mc_packet.Pool

type Client struct {
	conn             ReliableConnection
	close            chan struct{}
	byeSent          bool
	closed           bool
	pongDeadline     time.Time
	gamePackets      chan []byte
	uqHolderWaitChan chan []byte
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

type NoEOFByteReader struct {
	s []byte
	i int
}

func (nbr *NoEOFByteReader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	n = copy(b, nbr.s[nbr.i:])
	nbr.i += n
	return
}

func (nbr *NoEOFByteReader) ReadByte() (b byte, err error) {
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	b = nbr.s[nbr.i]
	nbr.i++
	return b, nil
}

func (c *Client) RecvDecodedGamePacket() (pk mc_packet.Packet, err error) {
	mcPkt, err := c.RecvGamePacket()
	if err != nil {
		return nil, err
	}
	pkFn, hasK := TypePool[uint32(mcPkt[0])]
	if !hasK {
		return nil, fmt.Errorf("cannot decode packet %v", mcPkt[0])
	}
	pk = pkFn()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			//bytes.NewReader().Read()
			pk.Unmarshal(protocol.NewReader(&NoEOFByteReader{s: mcPkt[1:]}, 0))
		}
	}()
	pk.Unmarshal(protocol.NewReader(&NoEOFByteReader{s: mcPkt[1:]}, 0))
	return pk, nil
}

// RequestUQHolder currently support request "*" only
func (c *Client) RequestUQHolder(request string) (*uqHolder.UQHolder, error) {
	err := c.Send(&packet.UQHolderRequestPacket{QueryString: []byte(request)})
	if err != nil {
		return nil, err
	}
	cont := <-c.uqHolderWaitChan
	fmt.Println(len(cont))
	uq := uqHolder.NewUQHolder(0)
	if err = uq.UnMarshal(cont); err != nil {
		return nil, err
	} else {
		return uq, nil
	}
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

func (c *Client) IsClosed() bool {
	return c.closed
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
				return
			}
			go func() { c.Send(&packet.PingPacket{}) }()
			//fmt.Println("Ping")
		case rawPacket := <-rc:
			c.pongDeadline = time.Now().Add(5 * time.Second)
			pkt, canParse := packet.Deserialize(rawPacket)
			if !canParse {
				fmt.Println(rawPacket)
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
			case *packet.UQHolderResponsePacket:
				c.uqHolderWaitChan <- p.Content
			}
		}
	}
}

func NewClient(address string) *Client {
	retChan := make(chan ReliableConnection)
	go func() {
		conn, err := KCPDial(address)
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
			conn:             conn,
			close:            make(chan struct{}),
			closed:           false,
			pongDeadline:     time.Now().Add(3 * time.Second),
			gamePackets:      make(chan []byte, 1024),
			uqHolderWaitChan: make(chan []byte),
		}
		if TypePool == nil {
			TypePool = mc_packet.NewPool()
		}
		go c.routine()
		return c
	}
}
