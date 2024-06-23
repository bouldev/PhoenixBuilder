/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type VConnection struct {
	Unknown1 int32
	Unknown2 []byte // Netease: uncertain type, read all
}

// ID ...
func (*VConnection) ID() uint32 {
	return IDVConnection
}

func (pk *VConnection) Marshal(io protocol.IO) {
	io.BEInt32(&pk.Unknown1)
	io.Bytes(&pk.Unknown2)
}
