// Netease
package protocol

// ------------------------- Enchant -------------------------

// 描述 物品堆栈 中的 单个 附魔属性
type Enchant struct {
	ID         uint16 `nbt:"id"`         // * TAG_Short(3) = 0
	Level      uint16 `nbt:"lvl"`        // * TAG_Short(3) = 0
	ModEnchant string `nbt:"modEnchant"` // TAG_String(8) = ""
}

func (e *Enchant) Marshal(r IO) {
	r.Uint16(&e.ID)
	r.Uint16(&e.Level)
	r.String(&e.ModEnchant)
}

func (e *Enchant) ToNBT() map[string]any {
	return map[string]any{
		"id":         int16(e.ID),
		"lvl":        int16(e.Level),
		"modEnchant": e.ModEnchant,
	}
}

func (e *Enchant) FromNBT(x map[string]any) {
	e.ID = uint16(x["id"].(int16))
	e.Level = uint16(x["lvl"].(int16))
	e.ModEnchant = x["modEnchant"].(string)
}

// ------------------------- EnchantList -------------------------

// 描述一个物品持有的多个 附魔 属性
type EnchantList struct {
	Data []Enchant // TAG_List[TAG_Compound] (9[10])
}

func (e *EnchantList) Marshal(r IO) {
	r.EnchantList(&e.Data)
}

func (e *EnchantList) ToNBT() []any {
	enchList := make([]any, len(e.Data))
	for key, value := range e.Data {
		enchList[key] = value.ToNBT()
	}
	return enchList
}

func (e *EnchantList) FromNBT(x []any) {
	e.Data = make([]Enchant, len(x))
	for key, value := range x {
		new := Enchant{}
		new.FromNBT(value.(map[string]any))
		e.Data[key] = new
	}
}
