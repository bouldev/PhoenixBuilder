package fields

import (
	"phoenixbuilder/minecraft/protocol"
)

// 描述 刷怪笼 中的一个复用字段
type SpawnData struct {
	Properties map[string]any `mapstructure:"Properties"` // TAG_Compound(10)
	TypeID     string         `mapstructure:"TypeId"`     // TAG_String(8)
	Weight     int32          `mapstructure:"Weight"`     // TAG_Int(4)
}

func (s *SpawnData) Marshal(r protocol.IO) {
	r.NBTWithLength(&s.Properties)
	r.String(&s.TypeID)
	r.Varint32(&s.Weight)
}
