/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

// Netease packet
type SetDimensionLocalTime struct {
	// The following fields are not used. (Added by Happy2018new)
	Unknown1 int32
	Unknown2 bool

	// Netease: Unknown1 and Unknown2 only,
	// but we read the whole content to this field to avoid panic
	Unknown3 []byte
}

// ID ...
func (*SetDimensionLocalTime) ID() uint32 {
	return IDSetDimensionLocalTime
}

func (pk *SetDimensionLocalTime) Marshal(io protocol.IO) {
	// if len > 0 {
	// 	io.Varint32(&pk.Unknown1)
	// 	io.Bool(&pk.Unknown2)
	// }
	io.Bytes(&pk.Unknown3)
}
