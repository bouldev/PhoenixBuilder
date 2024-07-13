package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 拼图方块
type JigsawBlock struct {
	general.BlockActor `mapstructure:",squash"`
	FinalState         string `mapstructure:"final_state"`        // TAG_String(8) = "minecraft:air"
	Joint              string `mapstructure:"joint"`              // TAG_String(8) = "rollable"
	Name               string `mapstructure:"name"`               // TAG_String(8) = "minecraft:empty"
	PlacementPriority  int32  `mapstructure:"placement_priority"` // Not used; TAG_Int(4) = 0
	SelectionPriority  int32  `mapstructure:"selection_priority"` // Not used; TAG_Int(4) = 0
	Target             string `mapstructure:"target"`             // TAG_String(8) = "minecraft:empty"
	TargetPool         string `mapstructure:"target_pool"`        // TAG_String(8) = "minecraft:empty"
}

// ID ...
func (*JigsawBlock) ID() string {
	return IDJigsawBlock
}

func (j *JigsawBlock) Marshal(io protocol.IO) {
	io.String(&j.Name)
	io.String(&j.Target)
	io.String(&j.TargetPool)
	io.String(&j.FinalState)
	io.String(&j.Joint)
	protocol.Single(io, &j.BlockActor)
}
