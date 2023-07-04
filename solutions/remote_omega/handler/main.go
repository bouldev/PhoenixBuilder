package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"phoenixbuilder/lib/encoding/binary_read_write"
	LE "phoenixbuilder/lib/encoding/little_endian"
	"phoenixbuilder/lib/minecraft/neomega/bundle"
	"phoenixbuilder/lib/minecraft/neomega/decouple/cmdsender"
	"phoenixbuilder/lib/minecraft/neomega/decouple/core"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/lib/minecraft/neomega/uqholder"
	"phoenixbuilder/lib/utils/sync_wrapper"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"

	zmq "github.com/go-zeromq/zmq4"
	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

type ZMQRpcCaller struct {
	socket zmq.Socket
	cbs    *sync_wrapper.SyncMap[func([][]byte)]
}

func NewZMQRpcCaller(ctx context.Context, endPoint string) (caller *ZMQRpcCaller, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var cancelFn func()
	ctx, cancelFn = context.WithCancel(ctx)
	id := zmq.SocketIdentity(fmt.Sprintf("dealer-%d", uuid.New().String()))
	socket := zmq.NewDealer(context.Background(), zmq.WithID(id))
	go func() {
		<-ctx.Done()
		socket.Close()
	}()
	if err = socket.Dial(endPoint); err != nil {
		cancelFn()
		return nil, fmt.Errorf("dialing: %w", err)
	}
	c := &ZMQRpcCaller{socket: socket, cbs: sync_wrapper.NewInstanceMap[func([][]byte)]()}
	go func() {
		var msg zmq.Msg
		for ctx.Err() == nil {
			msg, err = socket.Recv()
			if err != nil {
				pterm.Error.Println(err)
				break
			}
			idx := msg.Frames[0]
			rets := msg.Frames[1:]
			if cb, ok := c.cbs.Get(string(idx)); ok {
				cb(rets)
			}
		}
		cancelFn()
	}()
	return c, nil
}

func (c *ZMQRpcCaller) call(fn string, args [][]byte) (idx []byte) {
	idx = []byte(uuid.New().String())
	frames := append([][]byte{idx, []byte(fn)}, args...)
	err := c.socket.SendMulti(zmq.NewMsgFrom(frames...))
	if err != nil {
		return nil
	}
	return idx
}

func (c *ZMQRpcCaller) BlockCallAndGet(ctx context.Context, fn string, args [][]byte) (rets [][]byte) {
	if ctx == nil {
		ctx = context.Background()
	}
	idx := c.call(fn, args)
	if idx == nil {
		return nil
	}
	waitor := make(chan [][]byte)
	c.cbs.Set(string(idx), func(i [][]byte) {
		waitor <- i
	})
	select {
	case rets = <-waitor:
		return rets
	case <-ctx.Done():
		return nil
	}
}

func (c *ZMQRpcCaller) CallAndSetCB(fn string, args [][]byte, cb func(rets [][]byte)) {
	idx := c.call(fn, args)
	if idx != nil && cb != nil {
		c.cbs.Set(string(idx), cb)
	}
}

func (c *ZMQRpcCaller) CallNoResponse(fn string, args [][]byte) {
	c.call(fn, args)
}

func RevertToRawPacketWithShield(pk packet.Packet) (data []byte) {
	writer := bytes.NewBuffer(nil)
	packetID := pk.ID()
	LE.WriteInt32(binary_read_write.WrapBinaryWriter(writer), int32(packetID))
	w := protocol.NewWriter(writer, 0)
	pk.Marshal(w)
	return writer.Bytes()
}

func ConvertFromRawPacketWithShield(pool packet.Pool, data []byte) (pk packet.Packet) {
	reader := bytes.NewBuffer(data)
	packetID, _ := LE.Int32(binary_read_write.WrapBinaryReader(reader))
	r := protocol.NewReader(reader, 0)
	pk = pool[uint32(packetID)]()
	pk.Unmarshal(r)
	return pk
}

func main() {
	sub := zmq.NewSub(context.Background())
	defer sub.Close()
	err := sub.Dial("tcp://localhost:24016")
	if err != nil {
		log.Fatalf("could not dial: %v", err)
	}
	err = sub.SetOption(zmq.OptionSubscribe, "packet")
	if err != nil {
		log.Fatalf("could not subscribe: %v", err)
	}

	caller, err := NewZMQRpcCaller(nil, "tcp://localhost:24015")
	if err != nil {
		panic(err)
	}
	for {
		fmt.Println("waiting access point ready...")
		ret := caller.BlockCallAndGet(nil, "botReady", nil)
		if len(ret) > 0 && string(ret[0]) == "botReady" {
			break
		}
	}
	fmt.Println("access point ready")
	getUQHolder := func() omega.MicroUQHolder {
		uqHolderData := caller.BlockCallAndGet(nil, "getUQHolderBytes", nil)
		uq, err := uqholder.NewMicroUQHolderFromData(uqHolderData[0])
		if err != nil {
			panic(err)
		}
		return uq
	}
	//getShieldID := func() int32 {
	//	shieldIDBytes := caller.BlockCallAndGet(nil, "getConnShieldID", nil)
	//	return int32(binary.LittleEndian.Uint32(shieldIDBytes[0]))
	//}
	//packetIDBytes := make([]byte, 4)
	//sendPacketBytes := func(pktID uint32, data []byte) {
	//	binary.LittleEndian.PutUint32(packetIDBytes, pktID)
	//	caller.CallNoResponse("sendPacketBytes", [][]byte{packetIDBytes, data})
	//}
	sendPacket := func(pk packet.Packet) {
		caller.CallNoResponse("sendPacketBytes", [][]byte{RevertToRawPacketWithShield(pk)})
	}
	interactCore := core.NewRemoteInteractCore(sendPacket)
	//interactCore.UpdateShieldID(getShieldID())
	option := bundle.MicroOmegaOption{CmdSenderOptions: cmdsender.Options{ExpectedCmdFeedBack: false}}
	omegaCore := bundle.NewMicroOmega(interactCore, getUQHolder, option)

	go func() {
		//safeDecode := func(pkt packet.Packet, r *protocol.Reader) {
		//	defer func() {
		//		if recoveredErr := recover(); recoveredErr != nil {
		//			err = fmt.Errorf("%T: %w", pkt, recoveredErr.(error))
		//			//fmt.Println(err)
		//		}
		//	}()
		//	pkt.Unmarshal(r)
		//}
		//var pkt packet.Packet
		//var packetData []byte
		//var shieldID int32
		//var packetID uint32
		//shieldIDBytes := make([]byte, 4)
		//packetIDBytes = make([]byte, 4)
		//dataLenBytes := make([]byte, 4)
		pool := packet.NewPool()
		for {
			var msg zmq.Msg

			// Read envelope
			msg, err = sub.Recv()

			//if err != nil {
			//	panic(fmt.Errorf("could not receive message: %v", err))
			//}
			//shieldIDBytes, packetIDBytes, packetData, dataLenBytes = msg.Frames[1], msg.Frames[2], msg.Frames[3], msg.Frames[4]
			//shieldID = int32(binary.LittleEndian.Uint32(shieldIDBytes))
			//packetID = binary.LittleEndian.Uint32(packetIDBytes)
			//dataLen := binary.LittleEndian.Uint32(dataLenBytes)
			//if int(dataLen) != len(packetData) {
			//	fmt.Printf("%v!=%v\n", int(dataLen), len(packetData))
			//	continue
			//}
			//interactCore.UpdateShieldID(shieldID)
			//
			//reader := bytes.NewBuffer(packetData)
			//r := protocol.NewReader(reader, shieldID)
			//pkt = pool[packetID]()
			//safeDecode(pkt, r)
			omegaCore.HandlePacket(ConvertFromRawPacketWithShield(pool, msg.Frames[0]))
		}
	}()
	omegaCore.SendWSCmdAndInvokeOnResponse("tp 2401PT @s", func(output *packet.CommandOutput) {
		fmt.Println(output)
	})
	bufio.NewReader(os.Stdin).ReadByte()
}
