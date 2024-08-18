/*
PhoenixBuilder specific NEMC packet.
Author: Liliya233
*/
package packet

import "phoenixbuilder/minecraft/protocol"

// Netease packet
type ConfirmSkin struct {
	// Skin info
	SkinInfo []protocol.ConfirmSkinUnknownEntry
	// Launcher uids
	Uids []string
	// Netease
	Unknown3 []string
}

// ID ...
func (*ConfirmSkin) ID() uint32 {
	return IDConfirmSkin
}

func (pk *ConfirmSkin) Marshal(io protocol.IO) {
	protocol.SliceVaruint32Length(io, &pk.SkinInfo)
	protocol.FuncSliceOfLen(io, uint32(len(pk.SkinInfo)), &pk.Uids, io.String)
	protocol.FuncSliceOfLen(io, uint32(len(pk.SkinInfo)), &pk.Unknown3, io.String)
}
