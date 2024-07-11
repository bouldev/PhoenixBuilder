package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 可疑的方块
type BrushableBlock struct {
	general.BlockActor
	LootTable      string `nbt:"LootTable"`       // Not used; TAG_String(8) = "loot_tables/entities/empty_brushable_block.json"
	LootTableSeed  int32  `nbt:"LootTableSeed"`   // Not used; TAG_Int(4) = 0
	BrushCount     int32  `nbt:"brush_count"`     // Not used; TAG_Int(4) = 0
	BrushDirection byte   `nbt:"brush_direction"` // Not used; TAG_Byte(1) = 6
}

// ID ...
func (*BrushableBlock) ID() string {
	return IDBrushableBlock
}

func (b *BrushableBlock) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
}

func (b *BrushableBlock) ToNBT() map[string]any {
	return utils.MergeMaps(
		b.BlockActor.ToNBT(),
		map[string]any{
			"LootTable":       b.LootTable,
			"LootTableSeed":   b.LootTableSeed,
			"brush_count":     b.BrushCount,
			"brush_direction": b.BrushDirection,
		},
	)
}

func (b *BrushableBlock) FromNBT(x map[string]any) {
	b.BlockActor.FromNBT(x)
	b.LootTable = x["LootTable"].(string)
	b.LootTableSeed = x["LootTableSeed"].(int32)
	b.BrushCount = x["brush_count"].(int32)
	b.BrushDirection = x["brush_direction"].(byte)
}
