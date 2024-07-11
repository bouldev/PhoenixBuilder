package fields

import "phoenixbuilder/minecraft/protocol"

// 描述 旗帜 中所有的图案组合
type BannerPatterns struct {
	Color   int32  `nbt:"Color"`   // TAG_Int(4) = 0
	Pattern string `nbt:"Pattern"` // TAG_String(8) = ""
}

func (b *BannerPatterns) Marshal(r protocol.IO) {
	r.String(&b.Pattern)
	r.Varint32(&b.Color)
}

func (b *BannerPatterns) ToNBT() map[string]any {
	return map[string]any{
		"Color":   b.Color,
		"Pattern": b.Pattern,
	}
}

func (b *BannerPatterns) FromNBT(x map[string]any) {
	b.Color = x["Color"].(int32)
	b.Pattern = x["Pattern"].(string)
}
