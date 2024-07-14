package general

import (
	"phoenixbuilder/minecraft/protocol"
)

// 描述部分容器的 战利品表
type RandomizableBlockActor struct {
	BlockActor    `mapstructure:",squash"`
	LootTable     *string `mapstructure:"LootTable,omitempty"`     // TAG_String(8) = ""
	LootTableSeed int32   `mapstructure:"LootTableSeed,omitempty"` // TAG_Int(4) = 0
}

func (rba *RandomizableBlockActor) Marshal(r protocol.IO) {
	f := func() *string {
		if rba.LootTable == nil {
			rba.LootTable = new(string)
		}
		return rba.LootTable
	}

	protocol.NBTOptionalFunc(r, rba.LootTable, f, false, r.String)

	if rba.LootTable != nil && len(*rba.LootTable) > 0 {
		protocol.NBTInt(&rba.LootTableSeed, r.Varint64)
	} else {
		rba.LootTable = nil
	}

	protocol.Single(r, &rba.BlockActor)
}
