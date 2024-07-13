package general

import (
	"phoenixbuilder/minecraft/protocol"
)

// 描述 物品展示框 和 荧光物品展示框 的通用字段
type ItemFrameBlockActor struct {
	BlockActor     `mapstructure:",squash"`
	Item           *protocol.Item `mapstructure:"Item,omitempty"`           // TAG_Compound(10)
	ItemDropChance *float32       `mapstructure:"ItemDropChance,omitempty"` // TAG_Float(6) = 1
	ItemRotation   *float32       `mapstructure:"ItemRotation,omitempty"`   // TAG_Float(6) = 0
}

func (f *ItemFrameBlockActor) Marshal(r protocol.IO) {
	var has byte

	func1 := func() *protocol.Item {
		if f.Item == nil {
			f.Item = new(protocol.Item)
		}
		return f.Item
	}
	func2 := func() *float32 {
		if f.ItemDropChance == nil {
			f.ItemDropChance = new(float32)
		}
		return f.ItemDropChance
	}
	func3 := func() *float32 {
		if f.ItemRotation == nil {
			f.ItemRotation = new(float32)
		}
		return f.ItemRotation
	}

	if f.Item != nil && f.ItemDropChance != nil && f.ItemRotation != nil {
		has = 1
	}

	protocol.Single(r, &f.BlockActor)
	r.Uint8(&has)
	if has == 1 {
		protocol.NBTOptionalFunc(r, f.Item, func1, false, r.NBTItem)
		protocol.NBTOptionalFunc(r, f.ItemRotation, func3, false, r.Float32)
		protocol.NBTOptionalFunc(r, f.ItemDropChance, func2, false, r.Float32)
	}
}
