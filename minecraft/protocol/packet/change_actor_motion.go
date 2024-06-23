/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type ChangeActorMotion struct {
	Unknown1 int64
	Unknown2 uint8
}

// ID ...
func (*ChangeActorMotion) ID() uint32 {
	return IDChangeActorMotion
}

func (pk *ChangeActorMotion) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.Uint8(&pk.Unknown2)
}
