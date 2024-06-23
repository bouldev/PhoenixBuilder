/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type TransportNoCompress struct {
	Unknown1 int32
	Unknown2 []byte // Netease: uncertain type, read all
}

// ID ...
func (*TransportNoCompress) ID() uint32 {
	return IDTransportNoCompress
}

func (pk *TransportNoCompress) Marshal(io protocol.IO) {
	io.BEInt32(&pk.Unknown1)
	io.Bytes(&pk.Unknown2)
}
