/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type CustomV2 struct {
	Data []byte // Netease: uncertain type, read all
}

// ID ...
func (*CustomV2) ID() uint32 {
	return IDCustomV2
}

func (pk *CustomV2) Marshal(io protocol.IO) {
	io.Bytes(&pk.Data)
}
