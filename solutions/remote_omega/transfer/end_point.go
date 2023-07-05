package transfer

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	zmq "github.com/go-zeromq/zmq4"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/uqholder"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"
)

type Endpoint struct {
	sub    zmq.Socket
	caller *ZMQRpcCaller
	pool   packet.Pool
}

func NewEndPoint(ctx context.Context, pubAccessPoint, ctrlAccessPoint string) (endPoint *Endpoint, err error) {
	sub := zmq.NewSub(ctx)
	go func() {
		<-ctx.Done()
		sub.Close()
	}()
	err = sub.Dial(pubAccessPoint)
	if err != nil {
		return nil, err
	}
	err = sub.SetOption(zmq.OptionSubscribe, "packet")
	if err != nil {
		return nil, err
	}
	caller, err := NewZMQRpcCaller(nil, ctrlAccessPoint)
	if err != nil {
		panic(err)
	}
	return &Endpoint{
		pool:   packet.NewPool(),
		sub:    sub,
		caller: caller,
	}, nil
}

func (e *Endpoint) WaitReady() {
	for {
		if e.CheckAccessPointReady() {
			break
		}
		time.Sleep(time.Second / 10)
	}
}

func (e *Endpoint) CheckAccessPointReady() bool {
	ret := e.caller.BlockCallAndGet(nil, "botReady", nil)
	if len(ret) > 0 && string(ret[0]) == "botReady" {
		return true
	}
	return false
}

func (e *Endpoint) GetUQHolder() omega.MicroUQHolder {
	uqHolderData := e.caller.BlockCallAndGet(nil, "getUQHolderBytes", nil)
	uq, err := uqholder.NewMicroUQHolderFromData(uqHolderData[0])
	if err != nil {
		panic(err)
	}
	return uq
}

func (e *Endpoint) GetShieldID() int32 {
	shieldIDBytes := e.caller.BlockCallAndGet(nil, "getConnShieldID", nil)
	return int32(binary.LittleEndian.Uint32(shieldIDBytes[0]))
}

func (e *Endpoint) SendPacket(pk packet.Packet) {
	e.caller.CallNoResponse("sendPacket", [][]byte{RevertToRawPacketWithShield(pk)})
}

func (e *Endpoint) SendPacketData(pktID uint32, data []byte) {
	packetIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(packetIDBytes, pktID)
	e.caller.CallNoResponse("sendPacketBytes", [][]byte{packetIDBytes, data})
}

func (e *Endpoint) RecvPacket() (pk packet.Packet, shieldID int32, err error) {
	var msg zmq.Msg
	msg, err = e.sub.Recv()
	if err != nil {
		return nil, 0, err
	}
	shieldID = int32(binary.LittleEndian.Uint32(msg.Frames[2]))
	pk, err = ConvertFromRawPacketWithShield(e.pool, msg.Frames[1])
	if err != nil {
		return nil, 0, err
	}
	return pk, shieldID, err
}

func safeDecode(pkt packet.Packet, r *protocol.Reader) (p packet.Packet, err error) {
	defer func() {
		if recoveredErr := recover(); recoveredErr != nil {
			err = fmt.Errorf("%T: %w", pkt, recoveredErr.(error))
			//fmt.Println(err)
		}
	}()
	pkt.Unmarshal(r)
	return pkt, nil
}

func (e *Endpoint) RecvDirectPacket() (pk packet.Packet, shieldID int32, err error) {
	var msg zmq.Msg
	msg, err = e.sub.Recv()
	if err != nil {
		return nil, 0, err
	}
	shieldIDBytes, packetData, dataLenBytes := msg.Frames[1], msg.Frames[2], msg.Frames[3]
	shieldID = int32(binary.LittleEndian.Uint32(shieldIDBytes))
	dataLen := binary.LittleEndian.Uint32(dataLenBytes)
	if int(dataLen) != len(packetData) {
		return nil, 0, fmt.Errorf("len mismatch %v!=%v\n", int(dataLen), len(packetData))
	}
	reader := bytes.NewBuffer(packetData)
	header := &packet.Header{}
	if err := header.Read(reader); err != nil {
		return nil, 0, fmt.Errorf("error reading packet header: %v", err)
	}
	r := protocol.NewReader(reader, shieldID)
	if pktMake, found := e.pool[header.PacketID]; found {
		pk = pktMake()
	} else {
		return nil, 0, fmt.Errorf("pktID %v not found", header.PacketID)
	}
	pk, err = safeDecode(pk, r)
	if err != nil {
		return nil, 0, err
	}
	return pk, shieldID, nil
}
