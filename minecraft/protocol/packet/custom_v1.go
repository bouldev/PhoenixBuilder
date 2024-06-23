/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type CustomV1 struct {
	Data []byte // Netease: uncertain type, read all
}

// ID ...
func (*CustomV1) ID() uint32 {
	return IDCustomV1
}

func (pk *CustomV1) Marshal(io protocol.IO) {
	io.Bytes(&pk.Data)
}
