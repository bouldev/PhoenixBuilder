package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 活塞臂
type PistonArm struct {
	general.BlockActor `mapstructure:",squash"`
	AttachedBlocks     []any   `mapstructure:"AttachedBlocks"` // TAG_List[TAG_Int] (9[4])
	BreakBlocks        []any   `mapstructure:"BreakBlocks"`    // TAG_List[TAG_Int] (9[4])
	LastProgress       float32 `mapstructure:"LastProgress"`   // TAG_Float(6) = 0
	NewState           byte    `mapstructure:"NewState"`       // TAG_Byte(1) = 0
	Progress           float32 `mapstructure:"Progress"`       // TAG_Float(6) = 0
	State              byte    `mapstructure:"State"`          // TAG_Byte(1) = 0
	Sticky             byte    `mapstructure:"Sticky"`         // TAG_Byte(1) = 0
}

// ID ...
func (*PistonArm) ID() string {
	return IDPistonArm
}

func (p *PistonArm) Marshal(io protocol.IO) {
	protocol.Single(io, &p.BlockActor)
	io.Float32(&p.Progress)
	io.Float32(&p.LastProgress)
	protocol.NBTInt(&p.State, io.Varuint32)
	protocol.NBTInt(&p.NewState, io.Varuint32)
	io.Uint8(&p.Sticky)
	protocol.NBTSlice(io, &p.AttachedBlocks, func(t *[]int32) { io.PistonAttachedBlocks(t) })
	protocol.NBTSlice(io, &p.BreakBlocks, func(t *[]int32) { io.PistonAttachedBlocks(t) })
}
