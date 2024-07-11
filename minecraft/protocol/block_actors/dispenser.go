package block_actors

import general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"

// 发射器
type Dispenser struct {
	general.DispenserBlockActor
}

// ID ...
func (*Dispenser) ID() string {
	return IDDispenser
}
