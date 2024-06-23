/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type StoreBuySucc struct {
	Data []byte
}

// ID ...
func (*StoreBuySucc) ID() uint32 {
	return IDStoreBuySucc
}

func (pk *StoreBuySucc) Marshal(io protocol.IO) {
	io.ByteSlice(&pk.Data)
}
