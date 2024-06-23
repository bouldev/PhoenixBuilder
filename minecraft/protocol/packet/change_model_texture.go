/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type ChangeModelTexture struct {
	Unknown1 int64
	Unknown2 string
	Unknown3 int64
	Unknown4 uint8
}

// ID ...
func (*ChangeModelTexture) ID() uint32 {
	return IDChangeModelTexture
}

func (pk *ChangeModelTexture) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.String(&pk.Unknown2)
	io.Varint64(&pk.Unknown3)
	io.Uint8(&pk.Unknown4)
}
