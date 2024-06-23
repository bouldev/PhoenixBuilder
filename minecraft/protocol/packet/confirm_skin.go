/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type ConfirmSkin struct {
	Unknown1 []struct {
		Unknown1 bool
		Unknown2 uint64
		Unknown3 uint64
		Unknown4 string
	}
	Unknown2 []string
	Unknown3 []string
}

// ID ...
func (*ConfirmSkin) ID() uint32 {
	return IDConfirmSkin
}

// TODO: Make this func support marshal (Added by Happy2018new)
func (pk *ConfirmSkin) Marshal(io protocol.IO) {
	var count uint32
	io.Varuint32(&count)

	if count > 0 {
		pk.Unknown1 = make([]struct {
			Unknown1 bool
			Unknown2 uint64
			Unknown3 uint64
			Unknown4 string
		}, count)
		pk.Unknown2 = make([]string, count)
		pk.Unknown3 = make([]string, count)

		for i := uint32(0); i < count; i++ {
			io.Bool(&pk.Unknown1[i].Unknown1)
			io.Uint64(&pk.Unknown1[i].Unknown2)
			io.Uint64(&pk.Unknown1[i].Unknown3)
			io.String(&pk.Unknown1[i].Unknown4)
		}
		for i := uint32(0); i < count; i++ {
			io.String(&pk.Unknown2[i])
		}
		for i := uint32(0); i < count; i++ {
			io.String(&pk.Unknown3[i])
		}
	}
}
