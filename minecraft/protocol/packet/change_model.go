/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type ChangeModel struct {
	Unknown1 int64
	Unknown2 string
}

// ID ...
func (*ChangeModel) ID() uint32 {
	return IDChangeModel
}

func (pk *ChangeModel) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.String(&pk.Unknown2)
}
