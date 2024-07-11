package fields

import "phoenixbuilder/minecraft/protocol"

// 描述 告示牌 中的一个复用字段
type SignText struct {
	HideGlowOutline   byte   `nbt:"HideGlowOutline"`   // Not used; TAG_Byte(1) = 0
	IgnoreLighting    byte   `nbt:"IgnoreLighting"`    // TAG_Byte(1) = 0
	PersistFormatting byte   `nbt:"PersistFormatting"` // TAG_Byte(1) = 1
	SignTextColor     int32  `nbt:"SignTextColor"`     // TAG_Int(4) = -16777216
	Text              string `nbt:"Text"`              // TAG_String(8) = ""
	TextOwner         string `nbt:"TextOwner"`         // TAG_String(8) = ""
}

func (s *SignText) Marshal(r protocol.IO) {
	r.Uint8(&s.IgnoreLighting)
	r.Varint32(&s.SignTextColor)
	r.Uint8(&s.PersistFormatting)
	r.String(&s.TextOwner)
	r.String(&s.Text)
}

func (s *SignText) ToNBT() map[string]any {
	return map[string]any{
		"HideGlowOutline":   s.HideGlowOutline,
		"IgnoreLighting":    s.IgnoreLighting,
		"PersistFormatting": s.PersistFormatting,
		"SignTextColor":     s.SignTextColor,
		"Text":              s.Text,
		"TextOwner":         s.TextOwner,
	}
}

func (s *SignText) FromNBT(x map[string]any) {
	s.HideGlowOutline = x["HideGlowOutline"].(byte)
	s.IgnoreLighting = x["IgnoreLighting"].(byte)
	s.PersistFormatting = x["PersistFormatting"].(byte)
	s.SignTextColor = x["SignTextColor"].(int32)
	s.Text = x["Text"].(string)
	s.TextOwner = x["TextOwner"].(string)
}
