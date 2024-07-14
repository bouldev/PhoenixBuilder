/*
PhoenixBuilder specific fields.
Author: Happy2018new
*/
package protocol

import (
	"fmt"
	"phoenixbuilder/minecraft/nbt"

	"github.com/mitchellh/mapstructure"
)

// ------------------------- Item -------------------------

// 描述 物品堆栈 的额外字段，
// 如该物品的 tag 标签
type ExtraData map[string]any

// 描述一个 物品堆栈 实例
type Item struct {
	Count       byte           `mapstructure:"Count"`                // TAG_Byte(1) = 64
	Damage      int16          `mapstructure:"Damage"`               // TAG_Short(3) = 0
	Name        string         `mapstructure:"Name"`                 // TAG_String(8) = "minecraft:grass"
	WasPickedUp byte           `mapstructure:"WasPickedUp"`          // TAG_Byte(1) = 0
	Block       map[string]any `mapstructure:"Block,omitempty"`      // TAG_Compound(10)
	Tag         map[string]any `mapstructure:"tag,omitempty"`        // TAG_Compound(10)
	ModBlock    map[string]any `mapstructure:"modBlock,omitempty"`   // TAG_Compound(10)
	CanDestroy  []any          `mapstructure:"CanDestroy,omitempty"` // TAG_List(9)
	CanPlaceOn  []any          `mapstructure:"CanPlaceOn,omitempty"` // TAG_List(9)
}

// NBTItem reads a item data from the underlying buffer.
func (r *Reader) NBTItem(m *Item) {
	var extraData map[string]any
	var enchList []Enchant
	var nbtEnchList []map[string]any

	r.String(&m.Name)
	r.Uint8(&m.Count)
	r.Int16(&m.Damage)
	r.Uint8(&m.WasPickedUp)
	r.EnchantList(&enchList)
	r.NBT(&extraData, nbt.NetworkLittleEndian)

	for key, value := range extraData {
		switch key {
		case "Block":
			m.Block = value.(map[string]any)
		case "tag":
			m.Tag = value.(map[string]any)
		case "modBlock":
			m.ModBlock = value.(map[string]any)
		case "CanDestroy":
			m.CanDestroy = value.([]any)
		case "CanPlaceOn":
			m.CanPlaceOn = value.([]any)
		default:
			r.UnknownEnumOption(key, "NBT item extra data")
		}
	}

	if len(enchList) > 0 {
		if m.Tag == nil {
			m.Tag = make(map[string]any)
		}
		if err := mapstructure.Decode(enchList, &nbtEnchList); err != nil {
			panic(fmt.Sprintf("(r *Reader) NBTItem: %v", err))
		}
		new := make([]any, len(nbtEnchList))
		for key, value := range nbtEnchList {
			new[key] = value
		}
		m.Tag["ench"] = new
	}
}

// NBTItem writes a item data to the underlying buffer.
func (w *Writer) NBTItem(x *Item) {
	var enchList []Enchant
	extraData := make(map[string]any)

	if ench, hasEnch := x.Tag["ench"].([]any); hasEnch {
		err := mapstructure.Decode(ench, &enchList)
		if err != nil {
			panic(fmt.Sprintf("(w *Writer) NBTItem: %v", err))
		}
	}

	if len(x.Tag) > 0 {
		extraData["tag"] = make(map[string]any)
	}
	for key, value := range x.Tag {
		if key == "ench" {
			continue
		}
		extraData["tag"].(map[string]any)[key] = value
	}

	if len(x.Block) > 0 {
		extraData["Block"] = x.Block
	}
	if len(x.ModBlock) > 0 {
		extraData["modBlock"] = x.ModBlock
	}
	if len(x.CanDestroy) > 0 {
		extraData["CanDestroy"] = x.CanDestroy
	}
	if len(x.CanPlaceOn) > 0 {
		extraData["CanPlaceOn"] = x.CanPlaceOn
	}

	w.String(&x.Name)
	w.Uint8(&x.Count)
	w.Int16(&x.Damage)
	w.Uint8(&x.WasPickedUp)
	w.EnchantList(&enchList)
	w.NBT(&extraData, nbt.NetworkLittleEndian)
}

// ------------------------- ItemWithSlot -------------------------

// 描述一个 物品堆栈 实例，
// 但是同时负载该物品存放在容器中的槽位
type ItemWithSlot struct {
	Slot byte `mapstructure:"Slot"` // TAG_Byte(1) = 0
	Item `mapstructure:",squash"`
}

func (i *ItemWithSlot) Marshal(r IO) {
	r.Uint8(&i.Slot)
	r.NBTItem(&i.Item)
}
