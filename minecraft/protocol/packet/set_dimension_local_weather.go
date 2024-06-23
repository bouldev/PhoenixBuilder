/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import (
	"phoenixbuilder/minecraft/protocol"
)

// Netease packet
type SetDimensionLocalWeather struct {
	Unknown1 bool
	Unknown2 float32
	Unknown3 int32
	Unknown4 float32
	Unknown5 int32
	Unknown6 bool
}

// ID ...
func (*SetDimensionLocalWeather) ID() uint32 {
	return IDSetDimensionLocalWeather
}

func (pk *SetDimensionLocalWeather) Marshal(io protocol.IO) {
	io.Bool(&pk.Unknown1)
	io.Float32(&pk.Unknown2)
	io.Varint32(&pk.Unknown3)
	io.Float32(&pk.Unknown4)
	io.Varint32(&pk.Unknown5)
	io.Bool(&pk.Unknown6)
}
