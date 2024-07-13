package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 漏斗
type Hopper struct {
	general.BlockActor `mapstructure:",squash"`
	Items              []any `mapstructure:"Items"`            // TAG_List[TAG_Compound] (9[10])
	TransferCooldown   int32 `mapstructure:"TransferCooldown"` // TAG_Int(4) = 0
	MoveItemSpeed      int16 `mapstructure:"MoveItemSpeed"`    // TAG_Short(3) = 0
}

// ID ...
func (*Hopper) ID() string {
	return IDHopper
}

func (h *Hopper) Marshal(io protocol.IO) {
	protocol.Single(io, &h.BlockActor)
	protocol.NBTSlice(io, &h.Items, func(t *[]protocol.ItemWithSlot) { io.ItemList(t) })
	io.Varint32(&h.TransferCooldown)
	io.Varint16(&h.MoveItemSpeed)
}
