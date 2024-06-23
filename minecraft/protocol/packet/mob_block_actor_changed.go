/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease Packet
type MobBlockActorChanged struct {
	Position protocol.BlockPos
}

// ID ...
func (*MobBlockActorChanged) ID() uint32 {
	return IDMobBlockActorChanged
}

func (pk *MobBlockActorChanged) Marshal(io protocol.IO) {
	io.BlockPos(&pk.Position)
}
