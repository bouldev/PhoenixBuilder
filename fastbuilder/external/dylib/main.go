package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/external/connection"
	"phoenixbuilder/fastbuilder/external/packet"
	"phoenixbuilder/minecraft/protocol"
	mc_packet "phoenixbuilder/minecraft/protocol/packet"
	"time"
	"unsafe"

	"github.com/google/uuid"
)

var ErrSendOnClosedConnection = fmt.Errorf("send on closed connection")
var ErrRecvOnClosedConnection = fmt.Errorf("recv on closed connection")
var TypePool mc_packet.Pool

type Client struct {
	conn         connection.ReliableConnection
	close        chan struct{}
	byeSent      bool
	closed       bool
	pongDeadline time.Time
	gamePackets  chan []byte
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
	pk := TypePool[uint32(mcPkt[0])]()
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
			go func() { c.Send(&packet.PingPacket{}) }()
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
		}
		if TypePool == nil {
			TypePool = mc_packet.NewPool()
		}
		go c.routine()
		return c
	}
}

var DontGCMe []*Client

func objAvailable(id int) (*Client, error) {
	if id < 0 || id >= len(DontGCMe) {
		return nil, fmt.Errorf("id %v out of range %v\n", id, len(DontGCMe))
	}
	if c := DontGCMe[id]; c == nil {
		return nil, fmt.Errorf("id %v has been released", id)
	} else {
		return c, nil
	}
}

func toCErrStr(err error) *C.char {
	if err == nil {
		return nil
	}
	return C.CString(err.Error())
}

//export FreeMem
func FreeMem(address unsafe.Pointer) {
	C.free(address)
}

//export ConnectFB
func ConnectFB(address *C.char) (connID int, err *C.char) {
	str := C.GoString(address)
	// fmt.Println(str)
	client := NewClient(str)
	if client == nil {
		return -1, C.CString("connect fail")
	}
	for i, c := range DontGCMe {
		if c == nil {
			DontGCMe[i] = client
			return i, nil
		}
	}
	i := len(DontGCMe)
	DontGCMe = append(DontGCMe, client)
	return i, nil
}

//export ReleaseConnByID
func ReleaseConnByID(id int) (err *C.char) {
	if _, _err := objAvailable(id); _err != nil {
		return C.CString(_err.Error())
	} else {
		DontGCMe[id] = nil
	}
	for {
		if len(DontGCMe) != 0 && DontGCMe[len(DontGCMe)-1] == nil {
			DontGCMe = DontGCMe[0 : len(DontGCMe)-1]
		} else {
			break
		}
	}
	return nil
}

func bytesToCharArr(goByteSlice []byte) *C.char {
	ptr := C.malloc(C.size_t(len(goByteSlice)))
	C.memmove(ptr, (unsafe.Pointer)(&goByteSlice[0]), C.size_t(len(goByteSlice)))
	return (*C.char)(ptr)
}

//export RecvGamePacket
func RecvGamePacket(connID int) (pktBytes *C.char, l int, err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		return nil, 0, C.CString(_err.Error())
	}
	bs, _err := obj.RecvGamePacket()
	if _err != nil {
		bs = []byte{}
		ReleaseConnByID(connID)
		return nil, 0, C.CString(_err.Error())
	}
	//fmt.Println(bs)
	return bytesToCharArr(bs), len(bs), nil
}

//export SendGamePacketBytes
func SendGamePacketBytes(connID int, content []byte) (err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return C.CString(_err.Error())
	}
	_err = obj.Send(&packet.GamePacket{Content: content})
	return toCErrStr(_err)
}

//export SendFBCommand
func SendFBCommand(connID int, cmd *C.char) (err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return C.CString(_err.Error())
	}
	_err = obj.SendFBCmd(C.GoString(cmd))
	return toCErrStr(_err)
}

//export SendWSCommand
func SendWSCommand(connID int, cmd *C.char) (uuid *C.char, err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return nil, C.CString(_err.Error())
	}
	uid, _err := obj.SendWSCmd(C.GoString(cmd))
	return C.CString(uid.String()), toCErrStr(_err)
}

//export SendMCCommand
func SendMCCommand(connID int, cmd *C.char) (uuid *C.char, err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return nil, C.CString(_err.Error())
	}
	uid, _err := obj.SendMCCmd(C.GoString(cmd))
	return C.CString(uid.String()), toCErrStr(_err)
}

//export SendNoResponseCommand
func SendNoResponseCommand(connID int, cmd *C.char) (err *C.char) {
	obj, _err := objAvailable(connID)
	if _err != nil {
		ReleaseConnByID(connID)
		return C.CString(_err.Error())
	}
	_err = obj.SendNoResponseMCCmd(C.GoString(cmd))
	return toCErrStr(_err)
}

//export GamePacketBytesAsIsJsonStr
func GamePacketBytesAsIsJsonStr(pktBytes []byte) (jsonStr *C.char, err *C.char) {
	pk := TypePool[uint32(pktBytes[0])]()
	pk.Unmarshal(protocol.NewReader(bytes.NewReader(pktBytes[1:]), 0))
	marshal, _err := json.Marshal(pk)
	if _err != nil {
		return nil, C.CString(_err.Error())
	}
	return C.CString(string(marshal)), toCErrStr(_err)
}

//export JsonStrAsIsGamePacketBytes
func JsonStrAsIsGamePacketBytes(packetID int, jsonStr *C.char) (pktBytes *C.char, l int, err *C.char) {
	pk := TypePool[uint32(packetID)]()
	_err := json.Unmarshal([]byte(C.GoString(jsonStr)), &pk)
	if _err != nil {
		return nil, 0, C.CString(_err.Error())
	}
	b := &bytes.Buffer{}
	w := protocol.NewWriter(b, 0)
	hdr := pk.ID()
	w.Varuint32(&hdr)
	pk.Marshal(w)
	bs := b.Bytes()
	l = len(bs)
	return bytesToCharArr(bs), l, nil
}

func main() {
	//Windows: go build -o fbconn.dll -buildmode=c-shared main.go
	//Linux: go build -o libfbconn.so -buildmode=c-shared main.go
	//Macos: go build -o fbconn.dylib -buildmode=c-shared main.go
	//将生成的文件 (fbconn.dll 或 libfbconn.so 或 fbconn.dylib) 放在 conn.py 同一个目录下
}
