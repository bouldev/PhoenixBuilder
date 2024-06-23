/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease Packet
type AnimateEmoteEntity struct {
	Unknown1 string
	Unknown2 string
	Unknown3 string
	Unknown4 int32
	Unknown5 string
	Unknown6 float32
	Unknown7 []byte // Netease: uncertain, varint32 + slice = []sometype
}

// ID ...
func (*AnimateEmoteEntity) ID() uint32 {
	return IDAnimateEmoteEntity
}

func (pk *AnimateEmoteEntity) Marshal(io protocol.IO) {
	io.String(&pk.Unknown1)
	io.String(&pk.Unknown2)
	io.String(&pk.Unknown3)
	io.Int32(&pk.Unknown4)
	io.String(&pk.Unknown5)
	io.Float32(&pk.Unknown6)
	// io.ByteSlice(&pk.Unknown7)
	io.Bytes(&pk.Unknown7)
}
