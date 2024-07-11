package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 营火
type Campfire struct {
	general.BlockActor
	Item1     protocol.Optional[protocol.Item] `nbt:"Item1"`     // TAG_Compound(10)
	Item2     protocol.Optional[protocol.Item] `nbt:"Item2"`     // TAG_Compound(10)
	Item3     protocol.Optional[protocol.Item] `nbt:"Item3"`     // TAG_Compound(10)
	Item4     protocol.Optional[protocol.Item] `nbt:"Item4"`     // TAG_Compound(10)
	ItemTime1 int32                            `nbt:"ItemTime1"` // TAG_Int(4) = 0
	ItemTime2 int32                            `nbt:"ItemTime2"` // TAG_Int(4) = 0
	ItemTime3 int32                            `nbt:"ItemTime3"` // TAG_Int(4) = 0
	ItemTime4 int32                            `nbt:"ItemTime4"` // TAG_Int(4) = 0
}

// ID ...
func (*Campfire) ID() string {
	return IDCampfire
}

func (c *Campfire) Marshal(io protocol.IO) {
	protocol.Single(io, &c.BlockActor)
	io.Varint32(&c.ItemTime1)
	protocol.OptionalMarshaler(io, &c.Item1)
	io.Varint32(&c.ItemTime2)
	protocol.OptionalMarshaler(io, &c.Item2)
	io.Varint32(&c.ItemTime3)
	protocol.OptionalMarshaler(io, &c.Item3)
	io.Varint32(&c.ItemTime4)
	protocol.OptionalMarshaler(io, &c.Item4)
}

func (c *Campfire) ToNBT() map[string]any {
	temp := make(map[string]any)
	if item, has := c.Item1.Value(); has {
		temp = utils.MergeMaps(temp, map[string]any{"Item1": item.ToNBT()})
	}
	if item, has := c.Item2.Value(); has {
		temp = utils.MergeMaps(temp, map[string]any{"Item2": item.ToNBT()})
	}
	if item, has := c.Item3.Value(); has {
		temp = utils.MergeMaps(temp, map[string]any{"Item3": item.ToNBT()})
	}
	if item, has := c.Item4.Value(); has {
		temp = utils.MergeMaps(temp, map[string]any{"Item4": item.ToNBT()})
	}

	return utils.MergeMaps(
		c.BlockActor.ToNBT(),
		map[string]any{
			"ItemTime1": c.ItemTime1,
			"ItemTime2": c.ItemTime2,
			"ItemTime3": c.ItemTime3,
			"ItemTime4": c.ItemTime4,
		},
		temp,
	)
}

func (c *Campfire) FromNBT(x map[string]any) {
	c.BlockActor.FromNBT(x)
	c.ItemTime1 = x["ItemTime1"].(int32)
	c.ItemTime2 = x["ItemTime2"].(int32)
	c.ItemTime3 = x["ItemTime3"].(int32)
	c.ItemTime4 = x["ItemTime4"].(int32)

	if item, has := x["Item1"].(map[string]any); has {
		new := protocol.Item{}
		new.FromNBT(item)
		c.Item1 = protocol.Optional[protocol.Item]{Set: true, Val: new}
	}
	if item, has := x["Item2"].(map[string]any); has {
		new := protocol.Item{}
		new.FromNBT(item)
		c.Item2 = protocol.Optional[protocol.Item]{Set: true, Val: new}
	}
	if item, has := x["Item3"].(map[string]any); has {
		new := protocol.Item{}
		new.FromNBT(item)
		c.Item3 = protocol.Optional[protocol.Item]{Set: true, Val: new}
	}
	if item, has := x["Item4"].(map[string]any); has {
		new := protocol.Item{}
		new.FromNBT(item)
		c.Item4 = protocol.Optional[protocol.Item]{Set: true, Val: new}
	}
}
