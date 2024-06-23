/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type NeteaseJson struct {
	// Netease: string field,
	// but prased as byte slice for convenience
	Data []byte
}

// ID ...
func (*NeteaseJson) ID() uint32 {
	return IDNeteaseJson
}

func (pk *NeteaseJson) Marshal(io protocol.IO) {
	io.ByteSlice(&pk.Data)
}
