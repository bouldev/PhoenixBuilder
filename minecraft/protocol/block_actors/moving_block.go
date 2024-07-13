package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 移动的方块
type MovingBlock struct {
	general.BlockActor `mapstructure:",squash"`
	MovingBlock        map[string]any  `mapstructure:"movingBlock"`            // TAG_Compound(10)
	MovingBlockExtra   map[string]any  `mapstructure:"movingBlockExtra"`       // TAG_Compound(10)
	PistonPosX         int32           `mapstructure:"pistonPosX"`             // TAG_Int(4) = 0
	PistonPosY         int32           `mapstructure:"pistonPosY"`             // TAG_Int(4) = 0
	PistonPosZ         int32           `mapstructure:"pistonPosZ"`             // TAG_Int(4) = 0
	Expanding          byte            `mapstructure:"expanding"`              // Not used; TAG_Byte(1) = 0 or 1 (Boolean)
	MovingEntity       *map[string]any `mapstructure:"movingEntity,omitempty"` // TAG_Compound(10)
}

// ID ...
func (*MovingBlock) ID() string {
	return IDMovingBlock
}

func (m *MovingBlock) Marshal(io protocol.IO) {
	f := func() *map[string]any {
		if m.MovingEntity == nil {
			m.MovingEntity = new(map[string]any)
		}
		return m.MovingEntity
	}

	protocol.Single(io, &m.BlockActor)
	io.NBTWithLength(&m.MovingBlock)
	io.NBTWithLength(&m.MovingBlockExtra)
	io.Varint32(&m.PistonPosX)
	io.Varint32(&m.PistonPosY)
	io.Varint32(&m.PistonPosZ)
	protocol.NBTOptionalFunc(io, m.MovingEntity, f, true, io.NBTWithLength)
}
