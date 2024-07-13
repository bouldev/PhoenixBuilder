package block_actors

import general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"

// 幽匿感测体
type SculkSensor struct {
	general.BlockActor `mapstructure:",squash"`
}

// ID ...
func (*SculkSensor) ID() string {
	return IDSculkSensor
}
