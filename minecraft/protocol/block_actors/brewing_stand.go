package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 酿造台
type BrewingStand struct {
	general.BlockActor
	CookTime   int16             `nbt:"CookTime"`   // TAG_Short(3) = 0
	FuelAmount int16             `nbt:"FuelAmount"` // TAG_Short(3) = 0
	FuelTotal  int16             `nbt:"FuelTotal"`  // TAG_Short(3) = 0
	Items      protocol.ItemList `nbt:"Items"`      // TAG_List[TAG_Compound] (9[10])
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
	protocol.Single(io, &b.Items)
}

func (b *BrewingStand) ToNBT() map[string]any {
	return utils.MergeMaps(
		b.BlockActor.ToNBT(),
		map[string]any{
			"CookTime":   b.CookTime,
			"FuelAmount": b.FuelAmount,
			"FuelTotal":  b.FuelTotal,
			"Items":      b.Items.ToNBT(),
		},
	)
}

func (b *BrewingStand) FromNBT(x map[string]any) {
	b.BlockActor.FromNBT(x)
	b.CookTime = x["CookTime"].(int16)
	b.FuelAmount = x["FuelAmount"].(int16)
	b.FuelTotal = x["FuelTotal"].(int16)
	b.Items.FromNBT(x["Items"].([]any))
}
