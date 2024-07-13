package block_actors

import general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"

// 木桶
type Barrel struct {
	general.ChestBlockActor `mapstructure:",squash"`
}

// ID ...
func (*Barrel) ID() string {
	return IDBarrel
}
