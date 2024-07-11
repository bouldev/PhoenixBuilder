package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 活塞臂
type PistonArm struct {
	general.BlockActor
	AttachedBlocks []protocol.BlockPos `nbt:"AttachedBlocks"` // * TAG_List[TAG_Int] (9[4])
	BreakBlocks    []protocol.BlockPos `nbt:"BreakBlocks"`    // * TAG_List[TAG_Int] (9[4])
	LastProgress   float32             `nbt:"LastProgress"`   // TAG_Float(6) = 0
	NewState       uint32              `nbt:"NewState"`       // * TAG_Byte(1) = 0
	Progress       float32             `nbt:"Progress"`       // TAG_Float(6) = 0
	State          uint32              `nbt:"State"`          // * TAG_Byte(1) = 0
	Sticky         byte                `nbt:"Sticky"`         // TAG_Byte(1) = 0
}

// ID ...
func (*PistonArm) ID() string {
	return IDPistonArm
}

func (p *PistonArm) Marshal(io protocol.IO) {
	protocol.Single(io, &p.BlockActor)
	io.Float32(&p.Progress)
	io.Float32(&p.LastProgress)
	io.Varuint32(&p.State)
	io.Varuint32(&p.NewState)
	io.Uint8(&p.Sticky)
	protocol.FuncSliceVarint16Length(io, &p.AttachedBlocks, io.BlockPos)
	protocol.FuncSliceVarint16Length(io, &p.BreakBlocks, io.BlockPos)
}

func (p *PistonArm) ToNBT() map[string]any {
	attachedBlocks := make([]any, 0)
	for _, value := range p.AttachedBlocks {
		attachedBlocks = append(attachedBlocks, value[0], value[1], value[2])
	}
	breakBlocks := make([]any, 0)
	for _, value := range p.BreakBlocks {
		breakBlocks = append(breakBlocks, value[0], value[1], value[2])
	}
	return utils.MergeMaps(
		p.BlockActor.ToNBT(),
		map[string]any{
			"AttachedBlocks": attachedBlocks,
			"BreakBlocks":    breakBlocks,
			"LastProgress":   p.LastProgress,
			"NewState":       byte(p.NewState),
			"Progress":       p.Progress,
			"State":          byte(p.State),
			"Sticky":         p.Sticky,
		},
	)
}

func (p *PistonArm) FromNBT(x map[string]any) {
	p.BlockActor.FromNBT(x)
	attachedBlocks := utils.FromAnyList[int32](x["AttachedBlocks"].([]any))
	for i := 0; i < len(attachedBlocks)/3; i++ {
		index := i * 3
		p.AttachedBlocks = append(
			p.AttachedBlocks,
			protocol.BlockPos{attachedBlocks[index], attachedBlocks[index+1], attachedBlocks[index+2]},
		)
	}
	breakBlocks := utils.FromAnyList[int32](x["BreakBlocks"].([]any))
	for i := 0; i < len(breakBlocks)/3; i++ {
		index := i * 3
		p.BreakBlocks = append(
			p.BreakBlocks,
			protocol.BlockPos{breakBlocks[index], breakBlocks[index+1], breakBlocks[index+2]},
		)
	}
	p.LastProgress = x["LastProgress"].(float32)
	p.NewState = uint32(x["NewState"].(byte))
	p.Progress = x["Progress"].(float32)
	p.State = uint32(x["State"].(byte))
	p.Sticky = x["Sticky"].(byte)
}
