package transfer

import (
	"context"
	"encoding/binary"
	"fmt"
	zmq "github.com/go-zeromq/zmq4"
	"github.com/pterm/pterm"
	"os"
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
)

type TransferAccessPoint struct {
	pub         zmq.Socket
	ctrl        zmq.Socket
	getShieldID func() int32
}

func NewTransferAccessPoint(ctx context.Context, pubEndPoint, ctrlEndPoint string, getShieldID func() int32) (transfer *TransferAccessPoint, err error) {
	pub := zmq.NewPub(ctx)
	removeIPC := func(ipc string) error {
		if strings.HasPrefix(ipc, "ipc://") {
			ipcFile := ipc[len("ipc://"):]
			if _, err := os.Stat(ipcFile); err == nil {
				if err = os.Remove(ipcFile); err != nil {
					return err
				}
			}
		}
		return nil
	}
	err = removeIPC(pubEndPoint)
	if err != nil {
		return nil, err
	}
	err = pub.Listen(pubEndPoint)
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		pub.Close()
	}()
	ctrl := zmq.NewRouter(ctx, zmq.WithID(zmq.SocketIdentity("router")))
	go func() {
		<-ctx.Done()
		ctrl.Close()
	}()
	err = removeIPC(ctrlEndPoint)
	if err != nil {
		return nil, err
	}
	if err = ctrl.Listen(ctrlEndPoint); err != nil {
		return nil, fmt.Errorf("listening: %w", err)
	}
	if getShieldID == nil {
		getShieldID = func() int32 {
			return 0
		}
	}
	t := &TransferAccessPoint{
		pub,
		ctrl,
		getShieldID,
	}
	return t, nil
}

func (t *TransferAccessPoint) UpdateGetShieldIDFn(getShieldID func() int32) {
	t.getShieldID = getShieldID
}

func (t *TransferAccessPoint) PubGamePacket(pk packet.Packet) error {
	shieldIDBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(shieldIDBytes, uint32(t.getShieldID()))
	return t.pub.SendMulti(zmq.NewMsgFrom(
		[]byte("packet"),
		RevertToRawPacketWithShield(pk),
		shieldIDBytes,
	))
}

func (t *TransferAccessPoint) PubGamePacketData(packetData []byte) error {
	shieldIDBytes := make([]byte, 4)
	dataLenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(shieldIDBytes, uint32(t.getShieldID()))
	binary.LittleEndian.PutUint32(dataLenBytes, uint32(len(packetData)))
	return t.pub.SendMulti(zmq.NewMsgFrom(
		[]byte("packet"),
		shieldIDBytes,
		packetData,
		dataLenBytes,
	))
}

func (t *TransferAccessPoint) EnableCtrlTransfer(omega omega.MicroOmega) (err error) {
	getUQHolderBytes := func() []byte {
		data, err := omega.GetMicroUQHolder().Marshal()
		if err != nil {
			pterm.Error.Println(err)
		}
		return data
	}
	pool := packet.NewPool()
	return DefaultZMQFunctionSetsRpcExpose(t.ctrl, ZMQFunctionSets{
		"getUQHolderBytes": func(args [][]byte, onResult func([][]byte)) {
			onResult([][]byte{getUQHolderBytes()})
		},
		"sendPacket": func(args [][]byte, onResult func([][]byte)) {
			pk, err := ConvertFromRawPacketWithShield(pool, args[0])
			if err == nil {
				omega.GetGameControl().SendPacket(pk)
			}
		},
		"sendPacketBytes": func(args [][]byte, onResult func([][]byte)) {
			if len(args) == 2 {
				pktIDBytes := args[0]
				packetDataBytes := args[1]
				if len(pktIDBytes) == 4 {
					packetID := binary.LittleEndian.Uint32(pktIDBytes)
					omega.GetGameControl().SendPacketBytes(packetID, packetDataBytes)
				}
			}
		},
		"botReady": func(args [][]byte, onResult func([][]byte)) {
			onResult([][]byte{[]byte("botReady")})
		},
		"getConnShieldID": func(args [][]byte, onResult func([][]byte)) {
			shieldIDBytes := make([]byte, 4)
			binary.LittleEndian.PutUint32(shieldIDBytes, uint32(t.getShieldID()))
			onResult([][]byte{shieldIDBytes})
		},
	})
}
