package block_actors

import general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"

// 校频幽匿感测体
type CalibratedSculkSensor struct {
	general.BlockActor `mapstructure:",squash"`
}

// ID ...
func (*CalibratedSculkSensor) ID() string {
	return IDCalibratedSculkSensor
}
