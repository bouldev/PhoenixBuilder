package general

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/block_actors/fields"
)

// 描述各类告示牌的通用字段
type SignBlockActor struct {
	BlockActor
	BackText  fields.SignText
	FrontText fields.SignText
	IsWaxed   byte `nbt:"IsWaxed"` // TAG_Byte(1) = 0
}

func (s *SignBlockActor) Marshal(r protocol.IO) {
	protocol.Single(r, &s.BlockActor)
	protocol.Single(r, &s.FrontText)
	protocol.Single(r, &s.BackText)
	r.Uint8(&s.IsWaxed)
}

func (s *SignBlockActor) ToNBT() map[string]any {
	return utils.MergeMaps(
		s.BlockActor.ToNBT(),
		map[string]any{
			"BackText":  s.BackText.ToNBT(),
			"FrontText": s.FrontText.ToNBT(),
			"IsWaxed":   s.IsWaxed,
		},
	)
}

func (s *SignBlockActor) FromNBT(x map[string]any) {
	s.BlockActor.FromNBT(x)
	s.BackText.FromNBT(x["BackText"].(map[string]any))
	s.FrontText.FromNBT(x["FrontText"].(map[string]any))
	s.IsWaxed = x["IsWaxed"].(byte)
}
