package fields

import "phoenixbuilder/minecraft/protocol"

// 描述 化合物创建器 存放的物品
type ChemistryTableItem struct {
	ItemId    int32 `mapstructure:"itemId"`    // TAG_Int(4) = 0
	ItemAux   int16 `mapstructure:"itemAux"`   // TAG_Short(3) = 0
	ItemStack byte  `mapstructure:"itemStack"` // TAG_Byte(1) = 0
}

func (c *ChemistryTableItem) Marshal(r protocol.IO) {
	r.Varint32(&c.ItemId)
	r.Varint16(&c.ItemAux)
	protocol.NBTInt(&c.ItemStack, r.Varuint32)
}
