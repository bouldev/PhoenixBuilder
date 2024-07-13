package block_actors

import general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"

// 熔炉
type Furnace struct {
	general.FurnaceBlockActor `mapstructure:",squash"`
}

// ID ...
func (*Furnace) ID() string {
	return IDFurnace
}
