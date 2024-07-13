package general

import (
	"phoenixbuilder/minecraft/protocol"
)

// 描述 箱子、末影箱、木桶 和 潜影盒 的通用字段
type ChestBlockActor struct {
	RandomizableBlockActor `mapstructure:",squash"`

	Findable byte  `mapstructure:"Findable"` // TAG_Byte(1) = 0
	Items    []any `mapstructure:"Items"`    // TAG_List[TAG_Compound] (9[10])

	Pairlead    *byte  `mapstructure:"pairlead,omitempty"`    // TAG_Byte(1) = 0
	Pairx       *int32 `mapstructure:"pairx,omitempty"`       // TAG_Int(4) = 0
	Pairz       *int32 `mapstructure:"Pairz,omitempty"`       // TAG_Int(4) = 0
	ForceUnpair *byte  `mapstructure:"forceunpair,omitempty"` // TAG_Byte(1) = 1

	CustomSize *int16 `mapstructure:"CustomSize,omitempty"` // TAG_Short(3) = 0
}

func (c *ChestBlockActor) Marshal(r protocol.IO) {
	var hasPair byte

	func1 := func() *byte {
		if c.Pairlead == nil {
			c.Pairlead = new(byte)
		}
		return c.Pairlead
	}
	func2 := func() *int32 {
		if c.Pairx == nil {
			c.Pairx = new(int32)
		}
		return c.Pairx
	}
	func3 := func() *int32 {
		if c.Pairz == nil {
			c.Pairz = new(int32)
		}
		return c.Pairz
	}
	func4 := func() *byte {
		if c.ForceUnpair == nil {
			c.ForceUnpair = new(byte)
		}
		return c.ForceUnpair
	}
	func5 := func() *int16 {
		if c.CustomSize == nil {
			c.CustomSize = new(int16)
		}
		return c.CustomSize
	}

	if c.Pairx != nil && c.Pairz != nil {
		hasPair = 1
	}

	protocol.Single(r, &c.RandomizableBlockActor)
	protocol.NBTOptionalFunc(r, c.Pairlead, func1, false, r.Uint8)
	r.Uint8(&hasPair)

	if hasPair == 1 {
		protocol.NBTOptionalFunc(r, c.Pairx, func2, false, r.Varint32)
		protocol.NBTOptionalFunc(r, c.Pairz, func3, false, r.Varint32)
	} else {
		protocol.NBTOptionalFunc(r, c.ForceUnpair, func4, false, r.Uint8)
		if c.ForceUnpair != nil && *c.ForceUnpair != 1 {
			c.ForceUnpair = nil
		}
		c.Pairlead = nil
	}

	protocol.NBTOptionalFunc(r, c.CustomSize, func5, true, r.Varint16)
	protocol.NBTSlice(r, &c.Items, func(t *[]protocol.ItemWithSlot) { r.ItemList(t) })
	r.Uint8(&c.Findable)
}
