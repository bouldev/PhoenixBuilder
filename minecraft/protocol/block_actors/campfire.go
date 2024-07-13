package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 营火
type Campfire struct {
	general.BlockActor `mapstructure:",squash"`
	Item1              *protocol.Item `mapstructure:"Item1,omitempty"` // TAG_Compound(10)
	Item2              *protocol.Item `mapstructure:"Item2,omitempty"` // TAG_Compound(10)
	Item3              *protocol.Item `mapstructure:"Item3,omitempty"` // TAG_Compound(10)
	Item4              *protocol.Item `mapstructure:"Item4,omitempty"` // TAG_Compound(10)
	ItemTime1          int32          `mapstructure:"ItemTime1"`       // TAG_Int(4) = 0
	ItemTime2          int32          `mapstructure:"ItemTime2"`       // TAG_Int(4) = 0
	ItemTime3          int32          `mapstructure:"ItemTime3"`       // TAG_Int(4) = 0
	ItemTime4          int32          `mapstructure:"ItemTime4"`       // TAG_Int(4) = 0
}

// ID ...
func (*Campfire) ID() string {
	return IDCampfire
}

func (c *Campfire) Marshal(io protocol.IO) {
	func1 := func() *protocol.Item {
		if c.Item1 == nil {
			c.Item1 = new(protocol.Item)
		}
		return c.Item1
	}
	func2 := func() *protocol.Item {
		if c.Item2 == nil {
			c.Item2 = new(protocol.Item)
		}
		return c.Item2
	}
	func3 := func() *protocol.Item {
		if c.Item3 == nil {
			c.Item3 = new(protocol.Item)
		}
		return c.Item3
	}
	func4 := func() *protocol.Item {
		if c.Item4 == nil {
			c.Item4 = new(protocol.Item)
		}
		return c.Item4
	}

	protocol.Single(io, &c.BlockActor)
	io.Varint32(&c.ItemTime1)
	protocol.NBTOptionalFunc(io, c.Item1, func1, true, io.NBTItem)
	io.Varint32(&c.ItemTime2)
	protocol.NBTOptionalFunc(io, c.Item2, func2, true, io.NBTItem)
	io.Varint32(&c.ItemTime3)
	protocol.NBTOptionalFunc(io, c.Item3, func3, true, io.NBTItem)
	io.Varint32(&c.ItemTime4)
	protocol.NBTOptionalFunc(io, c.Item4, func4, true, io.NBTItem)
}
