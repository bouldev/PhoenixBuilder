package block_actors

import general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"

// 物品展示框
type ItemFrame struct {
	general.ItemFrameBlockActor `mapstructure:",squash"`
}

// ID ...
func (*ItemFrame) ID() string {
	return IDItemFrame
}
