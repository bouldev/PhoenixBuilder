package bot_privilege

import (
	"context"
	"encoding/json"
	"phoenixbuilder/fastbuilder/py_rpc"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
)

type PyRPCResponser struct {
	omega.MicroOmega
	isCheckNumResponded       bool
	chanCheckNumResponded     chan struct{}
	isGetStartTypeResponded   bool
	chanGetStartTypeResponded chan struct{}
	clientClosed              <-chan struct{}
	TransferData              func(content string, uid string) string
	TransferCheckNum          func(arg string) (ret string)
	Uid                       string
}

func NewPyRPCResponser(omega omega.MicroOmega, Uid string, clientClosed <-chan struct{}, TransferData func(content string, uid string) string, TransferCheckNum func(arg string) (ret string)) *PyRPCResponser {
	responser := &PyRPCResponser{
		MicroOmega:                omega,
		Uid:                       Uid,
		TransferData:              TransferData,
		TransferCheckNum:          TransferCheckNum,
		chanCheckNumResponded:     make(chan struct{}),
		chanGetStartTypeResponded: make(chan struct{}),
		clientClosed:              clientClosed,
	}
	omega.GetGameListener().SetOnTypedPacketCallBack(packet.IDPyRpc, responser.onPyRPC)
	return responser
}

func (o *PyRPCResponser) ChallengeCompete(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-o.clientClosed:
		return false
	case <-o.chanGetStartTypeResponded:
		if o.isCheckNumResponded {
			return true
		} else {
			return o.ChallengeCompete(ctx)
		}
	case <-o.chanCheckNumResponded:
		if o.isGetStartTypeResponded {
			return true
		} else {
			return o.ChallengeCompete(ctx)
		}
	}
}

func (o *PyRPCResponser) onPyRPC(pk packet.Packet) {
	pkt, ok := pk.(*packet.PyRpc)
	if !ok {
		return
	}
	goContentData := pkt.Value.MakeGo()
	content := goContentData.([]interface{})
	command := content[0].(string)
	data := content[1].([]interface{})
	if command == "S2CHeartBeat" {
		o.GetGameControl().SendPacket(&packet.PyRpc{
			Value: py_rpc.FromGo([]interface{}{
				"C2SHeartBeat",
				data,
				nil,
			}),
		})
	} else if command == "GetStartType" {
		response := o.TransferData(data[0].(string), o.Uid)
		o.GetGameControl().SendPacket(&packet.PyRpc{
			Value: py_rpc.FromGo([]interface{}{
				"SetStartType",
				[]interface{}{response},
				nil,
			}),
		})
		if !o.isGetStartTypeResponded {
			o.isGetStartTypeResponded = true
			close(o.chanGetStartTypeResponded)
		}
	} else if (command == "GetMCPCheckNum") && !o.isCheckNumResponded {
		firstArg := data[0].(string)
		secondArg := (data[1].([]interface{}))[0].(string)
		arg, _ := json.Marshal([]interface{}{firstArg, secondArg, o.GetMicroUQHolder().GetBotBasicInfo().GetBotUniqueID()})
		ret := o.TransferCheckNum(string(arg))
		ret_p := []interface{}{}
		json.Unmarshal([]byte(ret), &ret_p)
		o.GetGameControl().SendPacket(&packet.PyRpc{
			Value: py_rpc.FromGo([]interface{}{
				"SetMCPCheckNum",
				[]interface{}{
					ret_p,
				},
				nil,
			}),
		})
		o.isCheckNumResponded = true
		close(o.chanCheckNumResponded)
	}
}
