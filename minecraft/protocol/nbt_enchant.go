/*
PhoenixBuilder specific fields.
Author: Happy2018new
*/
package protocol

// 描述 物品堆栈 中的 单个 附魔属性
type Enchant struct {
	ID         int16  `mapstructure:"id"`         // TAG_Short(3) = 0
	Level      int16  `mapstructure:"lvl"`        // TAG_Short(3) = 0
	ModEnchant string `mapstructure:"modEnchant"` // TAG_String(8) = ""
}

func (e *Enchant) Marshal(r IO) {
	NBTInt(&e.ID, r.Uint16)
	NBTInt(&e.Level, r.Uint16)
	r.String(&e.ModEnchant)
}
