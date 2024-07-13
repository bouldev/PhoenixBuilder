package general

import (
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/block_actors/fields"
)

// 描述各类告示牌的通用字段
type SignBlockActor struct {
	BlockActor `mapstructure:",squash"`
	BackText   fields.SignText `mapstructure:"BackText"`  // TAG_Compound(10)
	FrontText  fields.SignText `mapstructure:"FrontText"` // TAG_Compound(10)
	IsWaxed    byte            `mapstructure:"IsWaxed"`   // TAG_Byte(1) = 0
}

func (s *SignBlockActor) Marshal(r protocol.IO) {
	protocol.Single(r, &s.BlockActor)
	protocol.Single(r, &s.FrontText)
	protocol.Single(r, &s.BackText)
	r.Uint8(&s.IsWaxed)
}
