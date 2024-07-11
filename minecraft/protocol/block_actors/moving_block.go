package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 移动的方块
type MovingBlock struct {
	general.BlockActor
	MovingBlock      map[string]any                    `nbt:"movingBlock"`      // TAG_Compound(10)
	MovingBlockExtra map[string]any                    `nbt:"movingBlockExtra"` // TAG_Compound(10)
	PistonPosX       int32                             `nbt:"pistonPosX"`       // TAG_Int(4) = 0
	PistonPosY       int32                             `nbt:"pistonPosY"`       // TAG_Int(4) = 0
	PistonPosZ       int32                             `nbt:"pistonPosZ"`       // TAG_Int(4) = 0
	Expanding        byte                              `nbt:"expanding"`        // Not used; TAG_Byte(1) = 0 or 1 (Boolean)
	MovingEntity     protocol.Optional[map[string]any] `nbt:"movingEntity"`     // TAG_Compound(10)
}

// ID ...
func (*MovingBlock) ID() string {
	return IDMovingBlock
}

func (m *MovingBlock) Marshal(io protocol.IO) {
	protocol.Single(io, &m.BlockActor)
	io.NBTWithLength(&m.MovingBlock)
	io.NBTWithLength(&m.MovingBlockExtra)
	io.Varint32(&m.PistonPosX)
	io.Varint32(&m.PistonPosY)
	io.Varint32(&m.PistonPosZ)
	protocol.OptionalFunc(io, &m.MovingEntity, io.NBTWithLength)
}

func (m *MovingBlock) ToNBT() map[string]any {
	var temp map[string]any
	if movingEntity, has := m.MovingEntity.Value(); has {
		temp = map[string]any{
			"movingEntity": movingEntity,
		}
	}
	return utils.MergeMaps(
		m.BlockActor.ToNBT(),
		map[string]any{
			"movingBlock":      m.MovingBlock,
			"movingBlockExtra": m.MovingBlockExtra,
			"pistonPosX":       m.PistonPosX,
			"pistonPosY":       m.PistonPosY,
			"pistonPosZ":       m.PistonPosZ,
			"expanding":        m.Expanding,
		},
		temp,
	)
}

func (m *MovingBlock) FromNBT(x map[string]any) {
	m.BlockActor.FromNBT(x)
	m.MovingBlock = x["movingBlock"].(map[string]any)
	m.MovingBlockExtra = x["movingBlockExtra"].(map[string]any)
	m.PistonPosX = x["pistonPosX"].(int32)
	m.PistonPosY = x["pistonPosY"].(int32)
	m.PistonPosZ = x["pistonPosZ"].(int32)
	m.Expanding = x["expanding"].(byte)

	if movingEntity, has := x["movingEntity"].(map[string]any); has {
		m.MovingEntity = protocol.Optional[map[string]any]{Set: true, Val: movingEntity}
	}
}
