package main

import (
	"bytes"
	"encoding/binary"
	"phoenixbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
)

type PyRPCResponser struct {
	omega.MicroOmega
	isCheckNumResponded bool
	TransferData        func(content string, uid string) string
	TransferCheckNum    func(firstArg string, secondArg string) (valM string, valS string)
	Uid                 string
}

func NewPyRPCResponser(omega omega.MicroOmega, Uid string, TransferData func(content string, uid string) string, TransferCheckNum func(firstArg string, secondArg string) (valM string, valS string)) *PyRPCResponser {
	responser := &PyRPCResponser{
		MicroOmega:       omega,
		Uid:              Uid,
		TransferData:     TransferData,
		TransferCheckNum: TransferCheckNum,
	}
	omega.GetGameListener().SetOnTypedPacketCallBack(packet.IDPyRpc, responser.onPyRPC)
	return responser
}

func (o *PyRPCResponser) onPyRPC(pk packet.Packet) {
	p, ok := pk.(*packet.PyRpc)
	if !ok {
		return
	}
	if strings.Contains(string(p.Content), "GetStartType") {
		// thank you rup!
		encData := p.Content[len(p.Content)-163 : len(p.Content)-1]
		response := o.TransferData(string(encData), o.Uid)
		//fmt.Printf("%s\n", response)
		o.GetGameControl().SendPacket(&packet.PyRpc{
			Content: bytes.Join([][]byte{[]byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xc, 0x53, 0x65, 0x74, 0x53, 0x74, 0x61, 0x72, 0x74, 0x54, 0x79, 0x70, 0x65, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0xc4},
				[]byte{byte(len(response))},
				[]byte(response),
				[]byte{0xc0},
			}, []byte{}),
		})

	} else if !o.isCheckNumResponded {
		if strings.Contains(string(p.Content), "GetMCPCheckNum") {
			// This shit sucks, so as netease.
			//fmt.Printf("%X", p.Content)
			//fmt.Printf("%s\n", p.Content)
			firstArgLenB := p.Content[19:21]
			firstArgLen := binary.BigEndian.Uint16(firstArgLenB)
			firstArg := string(p.Content[21 : 21+firstArgLen])
			secondArgLen := uint16(p.Content[23+firstArgLen])
			secondArg := string(p.Content[24+firstArgLen : 24+firstArgLen+secondArgLen])
			//fmt.Printf("%s\n", secondArg)
			//valM,_:=getUserInputMD5()
			//valS,_:=getUserInputMD5()
			//valM := utils.GetMD5(fmt.Sprintf("qhk+um%ssvdrx,9=>", secondArg))
			//valS := utils.GetMD5(fmt.Sprintf("%s%s", valM[16:], valM[:16]))
			//fmt.Printf("%s\n",valM)
			valM, valS := o.TransferCheckNum(firstArg, secondArg)
			/*conn.WritePacket(&packet.PyRpc{
				Content: bytes.Join([][]byte{[]byte{0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x93, 0xc4, 0xe, 0x53, 0x65, 0x74, 0x4d, 0x43, 0x50, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x82, 0xc4, 0x8, 0x5f, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x5f, 0xc4, 0x5, 0x74, 0x75, 0x70, 0x6c, 0x65, 0xc4, 0x5, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x91, 0xc4, 0x20},
					[]byte(valM),
					[]byte{0xc0},
				}, []byte{}),
			})*/
			o.GetGameControl().SendPacket(&packet.PyRpc{
				Content: bytes.Join([][]byte{[]byte{0x93, 0xc4, 0x0e}, []byte("SetMCPCheckNum"), []byte{0x91, 0x93, 0xc4, 0x20},
					[]byte(valM),
					[]byte{0xc4, 0x20},
					[]byte(valS),
					[]byte{0xc2},
					[]byte{0xC0},
				}, []byte{}),
			})
			o.isCheckNumResponded = true
		}

	}
}