package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 可疑的方块
type BrushableBlock struct {
	general.RandomizableBlockActor `mapstructure:",squash"`
	BrushCount                     int32 `mapstructure:"brush_count"`     // Not used; TAG_Int(4) = 0
	BrushDirection                 byte  `mapstructure:"brush_direction"` // Not used; TAG_Byte(1) = 6
}

// ID ...
func (*BrushableBlock) ID() string {
	return IDBrushableBlock
}

func (b *BrushableBlock) Marshal(io protocol.IO) {
	protocol.Single(io, &b.RandomizableBlockActor)
}
