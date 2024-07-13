package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 酿造台
type BrewingStand struct {
	general.BlockActor `mapstructure:",squash"`
	CookTime           int16 `mapstructure:"CookTime"`   // TAG_Short(3) = 0
	FuelAmount         int16 `mapstructure:"FuelAmount"` // TAG_Short(3) = 0
	FuelTotal          int16 `mapstructure:"FuelTotal"`  // TAG_Short(3) = 0
	Items              []any `mapstructure:"Items"`      // TAG_List[TAG_Compound] (9[10])
}

// ID ...
func (*BrewingStand) ID() string {
	return IDBrewingStand
}

func (b *BrewingStand) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
	io.Varint16(&b.FuelTotal)
	io.Varint16(&b.FuelAmount)
	io.Varint16(&b.CookTime)
	protocol.NBTSlice(io, &b.Items, func(t *[]protocol.ItemWithSlot) { io.ItemList(t) })
}
