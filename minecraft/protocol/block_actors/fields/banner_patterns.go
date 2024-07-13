package fields

import "phoenixbuilder/minecraft/protocol"

// 描述 旗帜 中所有的图案组合
type BannerPatterns struct {
	Color   int32  `mapstructure:"Color"`   // TAG_Int(4) = 0
	Pattern string `mapstructure:"Pattern"` // TAG_String(8) = ""
}

func (b *BannerPatterns) Marshal(r protocol.IO) {
	r.String(&b.Pattern)
	r.Varint32(&b.Color)
}
