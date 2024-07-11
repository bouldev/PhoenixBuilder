package fields

import "phoenixbuilder/minecraft/protocol"

// 描述 化合物创建器 存放的物品
type ChemistryTableItem struct {
	ItemId    int32  `nbt:"itemId"`    // TAG_Int(4) = 0
	ItemAux   int16  `nbt:"itemAux"`   // TAG_Short(3) = 0
	ItemStack uint32 `nbt:"itemStack"` // * TAG_Byte(1) = 0
}

// 检查 x 是否存在 ChemistryTableItem 中记录的所有数据
func (c *ChemistryTableItem) CheckExist(x map[string]any) (exist bool) {
	_, exist1 := x["itemId"]
	_, exist2 := x["itemAux"]
	_, exist3 := x["itemStack"]
	return exist1 && exist2 && exist3
}

func (c *ChemistryTableItem) Marshal(r protocol.IO) {
	r.Varint32(&c.ItemId)
	r.Varint16(&c.ItemAux)
	r.Varuint32(&c.ItemStack)
}

func (c *ChemistryTableItem) ToNBT() map[string]any {
	return map[string]any{
		"itemId":    c.ItemId,
		"itemAux":   c.ItemAux,
		"itemStack": byte(c.ItemStack),
	}
}

func (c *ChemistryTableItem) FromNBT(x map[string]any) {
	c.ItemId = x["itemId"].(int32)
	c.ItemAux = x["itemAux"].(int16)
	c.ItemStack = uint32(x["itemStack"].(byte))
}
