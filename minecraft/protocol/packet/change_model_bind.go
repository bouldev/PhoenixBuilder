/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type ChangeModelBind struct {
	Unknown1 int64
	Unknown2 int64
}

// ID ...
func (*ChangeModelBind) ID() uint32 {
	return IDChangeModelBind
}

func (pk *ChangeModelBind) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.Varint64(&pk.Unknown2)
}
