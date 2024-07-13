package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 炼药锅
type Cauldron struct {
	general.BlockActor `mapstructure:",squash"`
	Items              []any `mapstructure:"Items"`       // TAG_List[TAG_Compound] (9[10])
	PotionId           int16 `mapstructure:"PotionId"`    // TAG_Short(3) = -1
	PotionType         int16 `mapstructure:"PotionType"`  // TAG_Short(3) = -1
	CustomColor        int32 `mapstructure:"CustomColor"` // TAG_Int(4) = 0
}

// ID ...
func (*Cauldron) ID() string {
	return IDCauldron
}

func (c *Cauldron) Marshal(io protocol.IO) {
	protocol.Single(io, &c.BlockActor)
	protocol.NBTSlice(io, &c.Items, func(t *[]protocol.ItemWithSlot) { io.ItemList(t) })
	io.Varint16(&c.PotionId)
	io.Varint16(&c.PotionType)
	io.Varint32(&c.CustomColor)
}
