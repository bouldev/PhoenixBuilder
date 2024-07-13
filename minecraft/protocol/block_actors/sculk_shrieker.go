package block_actors

import general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"

// 幽匿尖啸体
type SculkShrieker struct {
	general.BlockActor `mapstructure:",squash"`
}

// ID ...
func (*SculkShrieker) ID() string {
	return IDSculkShrieker
}
