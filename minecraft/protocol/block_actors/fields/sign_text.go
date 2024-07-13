package fields

import "phoenixbuilder/minecraft/protocol"

// 描述 告示牌 中的一个复用字段
type SignText struct {
	HideGlowOutline   byte   `mapstructure:"HideGlowOutline"`   // Not used; TAG_Byte(1) = 0
	IgnoreLighting    byte   `mapstructure:"IgnoreLighting"`    // TAG_Byte(1) = 0
	PersistFormatting byte   `mapstructure:"PersistFormatting"` // TAG_Byte(1) = 1
	SignTextColor     int32  `mapstructure:"SignTextColor"`     // TAG_Int(4) = -16777216
	Text              string `mapstructure:"Text"`              // TAG_String(8) = ""
	TextOwner         string `mapstructure:"TextOwner"`         // TAG_String(8) = ""
}

func (s *SignText) Marshal(r protocol.IO) {
	r.Uint8(&s.IgnoreLighting)
	r.Varint32(&s.SignTextColor)
	r.Uint8(&s.PersistFormatting)
	r.String(&s.TextOwner)
	r.String(&s.Text)
}
