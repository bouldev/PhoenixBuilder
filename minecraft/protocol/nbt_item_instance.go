// Netease
package protocol

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/nbt"
)

// ------------------------- Item -------------------------

// 描述一个 物品堆栈 实例
type Item struct {
	Count       byte   `nbt:"Count"`       // TAG_Byte(1) = 64
	Damage      uint16 `nbt:"Damage"`      // * TAG_Short(3) = 0
	Name        string `nbt:"Name"`        // TAG_String(8) = "minecraft:grass"
	WasPickedUp byte   `nbt:"WasPickedUp"` // TAG_Byte(1) = 0

	ExtraData map[string]any // Modified from origin version, by NetEase; TAG_Compound(10)
	EnchList  EnchantList    // Modified from origin version, by NetEase
}

func (i *Item) Marshal(r IO) {
	r.String(&i.Name)
	r.Uint8(&i.Count)
	r.Varuint16(&i.Damage)
	r.Uint8(&i.WasPickedUp)
	i.EnchList.Marshal(r)
	r.NBT(&i.ExtraData, nbt.NetworkLittleEndian)
}

func (i *Item) ToNBT() map[string]any {
	temp := make(map[string]any)
	for key, value := range i.ExtraData {
		temp[key] = value
	}
	if len(i.EnchList.Data) > 0 {
		if _, has := temp["tag"]; !has {
			temp["tag"] = make(map[string]any)
		}
		temp["tag"].(map[string]any)["ench"] = i.EnchList.ToNBT()
	}
	return utils.MergeMaps(
		map[string]any{
			"Count":       i.Count,
			"Damage":      int16(i.Damage),
			"Name":        i.Name,
			"WasPickedUp": i.WasPickedUp,
		},
		temp,
	)
}

func (i *Item) FromNBT(x map[string]any) {
	i.Count = x["Count"].(byte)
	i.Damage = uint16(x["Damage"].(int16))
	i.Name = x["Name"].(string)
	i.WasPickedUp = x["WasPickedUp"].(byte)

	if tag, hasTag := x["tag"].(map[string]any); hasTag {
		if ench, hasEnch := tag["ench"].([]any); hasEnch {
			i.EnchList.FromNBT(ench)
		}
	}

	for key, value := range x {
		switch key {
		case "Count", "Damage", "Name", "WasPickedUp":
		default:
			if i.ExtraData == nil {
				i.ExtraData = make(map[string]any)
			}
			i.ExtraData[key] = value
			if key == "ench" {
				delete(i.ExtraData["tag"].(map[string]any), "ench")
			}
		}
	}
}

// ------------------------- ItemWithSlot -------------------------

// 描述一个 物品堆栈 实例，
// 但是同时负载该物品存放在容器中的槽位
type ItemWithSlot struct {
	Slot byte `nbt:"Slot"` // TAG_Byte(1) = 0
	Item
}

func (i *ItemWithSlot) Marshal(r IO) {
	r.Uint8(&i.Slot)
	i.Item.Marshal(r)
}

func (i *ItemWithSlot) ToNBT() map[string]any {
	return utils.MergeMaps(
		map[string]any{
			"Slot": i.Slot,
		},
		i.Item.ToNBT(),
	)
}

func (i *ItemWithSlot) FromNBT(x map[string]any) {
	i.Slot = x["Slot"].(byte)
	i.Item.FromNBT(x)
}

// ------------------------- ItemWithSlotList -------------------------

// 描述容器中存放的多个 物品堆栈 实例
type ItemList struct {
	Data []ItemWithSlot // TAG_List[TAG_Compound] (9[10])
}

func (i *ItemList) Marshal(r IO) {
	r.ItemList(&i.Data)
}

func (i *ItemList) ToNBT() []any {
	itemList := make([]any, len(i.Data))
	for key, value := range i.Data {
		itemList[key] = value.ToNBT()
	}
	return itemList
}

func (i *ItemList) FromNBT(x []any) {
	i.Data = make([]ItemWithSlot, len(x))
	for key, value := range x {
		new := ItemWithSlot{}
		new.FromNBT(value.(map[string]any))
		i.Data[key] = new
	}
}
