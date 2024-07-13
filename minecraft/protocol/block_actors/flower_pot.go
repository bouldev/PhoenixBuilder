package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 花盆
type FlowerPot struct {
	general.BlockActor `mapstructure:",squash"`
	PlantBlock         map[string]any `mapstructure:"PlantBlock"` // TAG_Compound(10)
}

// ID ...
func (*FlowerPot) ID() string {
	return IDFlowerPot
}

func (f *FlowerPot) Marshal(io protocol.IO) {
	protocol.Single(io, &f.BlockActor)
	io.NBTWithLength(&f.PlantBlock)
}
