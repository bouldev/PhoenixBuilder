/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type ChangeModelOffset struct {
	Unknown1 int64
	Unknown2 []byte // Netease: uncertain, 2 same operations
}

// ID ...
func (*ChangeModelOffset) ID() uint32 {
	return IDChangeModelOffset
}

func (pk *ChangeModelOffset) Marshal(io protocol.IO) {
	io.Varint64(&pk.Unknown1)
	io.Bytes(&pk.Unknown2)
}
