/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type HungerAttr struct {
	Data []byte // Netease: uncertain type, read all
}

// ID ...
func (*HungerAttr) ID() uint32 {
	return IDHungerAttr
}

func (pk *HungerAttr) Marshal(io protocol.IO) {
	io.Bytes(&pk.Data)
}
