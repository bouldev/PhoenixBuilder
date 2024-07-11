package general

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
)

// 描述 箱子、末影箱、木桶 和 潜影盒 的通用字段
type ChestBlockActor struct {
	RandomizableBlockActor
	Findable    byte                     `nbt:"Findable"` // TAG_Byte(1) = 0
	Items       protocol.ItemList        `nbt:"Items"`    // TAG_List[TAG_Compound] (9[10])
	HasPair     byte                     // Not a TAG, but a mark used to decide how to decode these four fields.
	Pairlead    byte                     `nbt:"pairlead"`    // TAG_Byte(1) = 0
	Pairx       int32                    `nbt:"pairx"`       // TAG_Int(4) = 0
	Pairz       int32                    `nbt:"Pairz"`       // TAG_Int(4) = 0
	ForceUnpair byte                     `nbt:"forceunpair"` // TAG_Byte(1) = 1
	CustomSize  protocol.Optional[int16] `nbt:"CustomSize"`  // TAG_Short(3) = 0
}

func (c *ChestBlockActor) Marshal(r protocol.IO) {
	protocol.Single(r, &c.RandomizableBlockActor)

	r.Uint8(&c.Pairlead)
	r.Uint8(&c.HasPair)
	if c.HasPair == 1 {
		r.Varint32(&c.Pairx)
		r.Varint32(&c.Pairz)
	} else {
		r.Uint8(&c.ForceUnpair)
	}

	protocol.OptionalFunc(r, &c.CustomSize, r.Varint16)
	protocol.Single(r, &c.Items)
	r.Uint8(&c.Findable)
}

func (c *ChestBlockActor) ToNBT() map[string]any {
	var pair map[string]any
	var customSize map[string]any

	if c.HasPair == 1 {
		pair = map[string]any{
			"pairlead": c.Pairlead,
			"pairx":    c.Pairx,
			"pairz":    c.Pairz,
		}
	} else if c.ForceUnpair == 1 {
		pair = map[string]any{
			"forceunpair": c.ForceUnpair,
		}
	}

	if data, has := c.CustomSize.Value(); has {
		customSize = map[string]any{
			"CustomSize": data,
		}
	}

	return utils.MergeMaps(
		c.RandomizableBlockActor.ToNBT(),
		map[string]any{
			"Findable": c.Findable,
			"Items":    c.Items.ToNBT(),
		},
		pair, customSize,
	)
}

func (c *ChestBlockActor) FromNBT(x map[string]any) {
	c.RandomizableBlockActor.FromNBT(x)
	c.Findable = x["Findable"].(byte)
	c.Items.FromNBT(x["Items"].([]any))

	if pairlead, has := x["pairlead"].(byte); has {
		c.HasPair = 1
		c.Pairlead = pairlead
		c.Pairx = x["pairx"].(int32)
		c.Pairz = x["pairz"].(int32)
	} else {
		c.ForceUnpair, _ = x["ForceUnpair"].(byte)
	}
	if customSize, has := x["CustomSize"].(int16); has {
		c.CustomSize = protocol.Optional[int16]{Set: true, Val: customSize}
	}
}
