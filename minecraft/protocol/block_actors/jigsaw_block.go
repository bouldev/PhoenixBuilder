package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 拼图方块
type JigsawBlock struct {
	general.BlockActor
	FinalState        string `nbt:"final_state"`        // TAG_String(8) = "minecraft:air"
	Joint             string `nbt:"joint"`              // TAG_String(8) = "rollable"
	Name              string `nbt:"name"`               // TAG_String(8) = "minecraft:empty"
	PlacementPriority int32  `nbt:"placement_priority"` // Not used; TAG_Int(4) = 0
	SelectionPriority int32  `nbt:"selection_priority"` // Not used; TAG_Int(4) = 0
	Target            string `nbt:"target"`             // TAG_String(8) = "minecraft:empty"
	TargetPool        string `nbt:"target_pool"`        // TAG_String(8) = "minecraft:empty"
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

func (j *JigsawBlock) ToNBT() map[string]any {
	return utils.MergeMaps(
		map[string]any{
			"final_state":        j.FinalState,
			"joint":              j.Joint,
			"name":               j.Name,
			"placement_priority": j.PlacementPriority,
			"selection_priority": j.SelectionPriority,
			"target":             j.Target,
			"target_pool":        j.TargetPool,
		},
		j.BlockActor.ToNBT(),
	)
}

func (j *JigsawBlock) FromNBT(x map[string]any) {
	j.FinalState = x["final_state"].(string)
	j.Joint = x["joint"].(string)
	j.Name = x["name"].(string)
	j.PlacementPriority = x["placement_priority"].(int32)
	j.SelectionPriority = x["selection_priority"].(int32)
	j.Target = x["target"].(string)
	j.TargetPool = x["target_pool"].(string)
	j.BlockActor.FromNBT(x)
}
