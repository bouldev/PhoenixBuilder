package main

import (
	"bytes"
	"context"
	"fmt"
	zmq "github.com/go-zeromq/zmq4"
	"github.com/pterm/pterm"
	"log"
	"phoenixbuilder/lib/encoding/binary_read_write"
	LE "phoenixbuilder/lib/encoding/little_endian"
	"phoenixbuilder/lib/helpers/fb_enter_server"
	"phoenixbuilder/lib/helpers/fbuser"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

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
	authServer := "wss://api.fastbuilder.pro:2053/"
	fmt.Println("Reading Info...")
	userName, userPassword, userToken, serverCode, serverPassword, err := fbuser.ReadInfo("", "", "", "", "")
	if err != nil {
		panic(err)
	}

	accessOption := fb_enter_server.MakeDefaultOption()
	accessOption.AuthServer = authServer
	accessOption.FBUserName = userName
	accessOption.FBUserPassword = userPassword
	accessOption.FBUserToken = userToken
	accessOption.ServerCode = serverCode
	accessOption.ServerPassword = serverPassword
	accessOption.MakeBotCreative = true
	accessOption.DisableCommandBlock = false
	accessOption.ExpectedCmdFeedBack = false

	pub := zmq.NewPub(context.Background())
	defer pub.Close()

	err = pub.Listen("tcp://*:24016")
	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}

	//var getShieldID func() int32
	accessOption.ReadLoopFunction = func(conn *minecraft.Conn, deadReason chan<- error, omega omega.ReactCore) {
		var pkt packet.Packet
		//var err error
		//var shieldID int32
		//packetData := []byte{}
		//shieldIDBytes := make([]byte, 4)
		//packetIDBytes := make([]byte, 4)
		//getShieldID = conn.GetShieldID
		//dataLenBytes := make([]byte, 4)
		for {

			pkt, err = conn.ReadPacket()
			//packetData = packetData[:]
			if err != nil {
				deadReason <- fmt.Errorf("%v: %v", fb_enter_server.ErrRentalServerDisconnected, err)
			}
			//shieldID = conn.GetShieldID()
			//binary.LittleEndian.PutUint32(shieldIDBytes, uint32(shieldID))
			//binary.LittleEndian.PutUint32(packetIDBytes, uint32(pkt.ID()))
			//binary.LittleEndian.PutUint32(dataLenBytes, uint32(len(packetData)))
			err = pub.SendMulti(zmq.NewMsgFrom(
				//[]byte("packet"),
				//shieldIDBytes[:],
				//packetIDBytes[:],
				//packetData[:],
				//dataLenBytes[:],
				RevertToRawPacketWithShield(pkt),
			))
			if err != nil {
				deadReason <- fmt.Errorf("fail to remote dispatch packets")
			}
			omega.HandlePacket(pkt)
		}

	}

	var deadReason chan error
	var omegaCore omega.MicroOmega
	omegaCore, deadReason, err = fb_enter_server.AccessServer(nil, accessOption)
	if err != nil {
		panic(err)
	}
	getUQHolderBytes := func() []byte {
		data, err := omegaCore.GetMicroUQHolder().Marshal()
		if err != nil {
			pterm.Error.Println(err)
		}
		return data
	}
	//sendPacketBytes := omegaCore.GetGameControl().SendPacketBytes
	pool := packet.NewPool()

	go func() {
		err := DefaultZMQFunctionSetsRpcExpose(nil, ZMQFunctionSets{
			"getUQHolderBytes": func(args [][]byte, onResult func([][]byte)) {
				onResult([][]byte{getUQHolderBytes()})
			},
			"sendPacketBytes": func(args [][]byte, onResult func([][]byte)) {
				omegaCore.GetGameControl().SendPacket(ConvertFromRawPacketWithShield(pool, args[0]))
			},
			"botReady": func(args [][]byte, onResult func([][]byte)) {
				onResult([][]byte{[]byte("botReady")})
			},
			//"getConnShieldID": func(args [][]byte, onResult func([][]byte)) {
			//	shieldIDBytes := make([]byte, 4)
			//	binary.LittleEndian.PutUint32(shieldIDBytes, uint32(getShieldID()))
			//	onResult([][]byte{shieldIDBytes})
			//},
		}, "tcp://*:24015")
		if err != nil {
			panic(err)
		}
	}()
	fmt.Println("bot ready")

	err = <-deadReason
	panic(err)
}

type ZMQFunctionSets map[string]func(args [][]byte, onResult func([][]byte))

func DefaultZMQFunctionSetsRpcExpose(ctx context.Context, sets ZMQFunctionSets, endPoint string) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	socket := zmq.NewRouter(ctx, zmq.WithID(zmq.SocketIdentity("router")))
	defer socket.Close()
	if err = socket.Listen(endPoint); err != nil {
		return fmt.Errorf("listening: %w", err)
	}
	//ErrCallFormat := zmq.NewMsgFrom([]byte("err"), []byte("illegal frames length"))
	//ErrCallFnName := zmq.NewMsgFrom([]byte("err"), []byte("function name not found"))
	for {
		var msg zmq.Msg
		msg, err = socket.Recv()
		if err != nil {
			return fmt.Errorf("receiving: %w", err)
		}
		frames := msg.Frames
		if len(frames) < 3 {
			//socket.Send(ErrCallFormat)
			continue
		}
		fnName := string(frames[2])
		if fn, found := sets[fnName]; found {
			fn(frames[3:], func(rets [][]byte) {
				socket.SendMulti(zmq.NewMsgFrom(append(frames[:2], rets...)...))
			})
		} else {
			//socket.Send(ErrCallFnName)
			continue
		}
	}
}
