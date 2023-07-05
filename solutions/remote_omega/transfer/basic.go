package transfer

import (
	"bytes"
	"context"
	"fmt"
	zmq "github.com/go-zeromq/zmq4"
	"github.com/google/uuid"
	"github.com/pterm/pterm"
	"phoenixbuilder/fastbuilder/lib/encoding/binary_read_write"
	LE "phoenixbuilder/fastbuilder/lib/encoding/little_endian"
	"phoenixbuilder/fastbuilder/lib/utils/sync_wrapper"
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

func ConvertFromRawPacketWithShield(pool packet.Pool, data []byte) (pk packet.Packet, err error) {
	reader := bytes.NewBuffer(data)
	packetID, _ := LE.Int32(binary_read_write.WrapBinaryReader(reader))
	r := protocol.NewReader(reader, 0)
	if pkMaker, found := pool[uint32(packetID)]; found {
		pk = pkMaker()
	} else {
		return nil, fmt.Errorf("pktID %v not found", uint32(packetID))
	}
	pk.Unmarshal(r)
	return pk, nil
}

type ZMQFunctionSets map[string]func(args [][]byte, onResult func([][]byte))

func DefaultZMQFunctionSetsRpcExpose(socket zmq.Socket, sets ZMQFunctionSets) (err error) {
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
