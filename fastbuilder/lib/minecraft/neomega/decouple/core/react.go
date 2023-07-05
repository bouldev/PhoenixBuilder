package core

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/neomega/omega"
	"phoenixbuilder/minecraft/protocol/packet"
)

func init() {
	if false {
		func(omega.ReactCore) {}(&ReactCore{})
	}
}

type ReactCore struct {
	onAnyPacketCallBack    []func(packet.Packet)
	onTypedPacketCallBacks map[uint32][]func(packet.Packet)
}

func NewReactCore() omega.ReactCore {
	return &ReactCore{
		onAnyPacketCallBack:    make([]func(packet2 packet.Packet), 0),
		onTypedPacketCallBacks: make(map[uint32][]func(packet.Packet)),
	}
}

func (r *ReactCore) SetOnAnyPacketCallBack(cb func(packet.Packet)) {
	r.onAnyPacketCallBack = append(r.onAnyPacketCallBack, cb)
}

func (r *ReactCore) SetOnTypedPacketCallBack(pktID uint32, cb func(packet.Packet)) {
	if _, ok := r.onTypedPacketCallBacks[pktID]; !ok {
		r.onTypedPacketCallBacks[pktID] = make([]func(packet2 packet.Packet), 0, 1)
	}
	r.onTypedPacketCallBacks[pktID] = append(r.onTypedPacketCallBacks[pktID], cb)
}

func (r *ReactCore) HandlePacket(pkt packet.Packet) {
	pktID := pkt.ID()
	// if pktID == packet.IDCommandOutput {
	// 	s, _ := json.Marshal(pkt)
	// 	fmt.Println(string(s))
	// }
	for _, cb := range r.onAnyPacketCallBack {
		cb(pkt)
	}
	if cbs, ok := r.onTypedPacketCallBacks[pktID]; ok {
		for _, cb := range cbs {
			cb(pkt)
		}
	}
}
