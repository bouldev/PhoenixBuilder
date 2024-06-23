/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type Transport struct {
	Unknown1 int32
	Unknown2 []byte // Netease: uncertain type, read all
}

// ID ...
func (*Transport) ID() uint32 {
	return IDTransport
}

func (pk *Transport) Marshal(io protocol.IO) {
	io.BEInt32(&pk.Unknown1)
	io.Bytes(&pk.Unknown2)
}
