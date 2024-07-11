package general

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
)

// 描述部分容器的 战利品表
type RandomizableBlockActor struct {
	BlockActor
	LootTable     string `nbt:"LootTable"`     // TAG_String(8) = ""
	LootTableSeed int64  `nbt:"LootTableSeed"` // * TAG_Int(4) = 0
}

func (l *RandomizableBlockActor) Marshal(r protocol.IO) {
	r.String(&l.LootTable)
	if len(l.LootTable) > 0 {
		r.Varint64(&l.LootTableSeed)
	}
	protocol.Single(r, &l.BlockActor)
}

func (l *RandomizableBlockActor) ToNBT() map[string]any {
	if len(l.LootTable) > 0 {
		return utils.MergeMaps(
			l.BlockActor.ToNBT(),
			map[string]any{
				"LootTable":     l.LootTable,
				"LootTableSeed": int32(l.LootTableSeed),
			},
		)
	}
	return l.BlockActor.ToNBT()
}

func (l *RandomizableBlockActor) FromNBT(x map[string]any) {
	if lootTable, has := x["LootTable"].(string); has {
		l.LootTable = lootTable
		l.LootTableSeed = int64(x["LootTableSeed"].(int32))
	}
	l.BlockActor.FromNBT(x)
}
