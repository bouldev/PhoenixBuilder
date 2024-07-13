package block_actors

import general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"

// 幽匿催发体
type SculkCatalyst struct {
	general.BlockActor `mapstructure:",squash"`
}

// ID ...
func (*SculkCatalyst) ID() string {
	return IDSculkCatalyst
}
