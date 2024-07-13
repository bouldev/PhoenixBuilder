package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/block_actors/fields"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 旗帜
type Banner struct {
	general.BlockActor `mapstructure:",squash"`
	Base               int32 `mapstructure:"Base"`               // TAG_Int(4) = 0
	Patterns           []any `mapstructure:"Patterns,omitempty"` // TAG_List[TAG_Compound] (9[10])
	Type               int32 `mapstructure:"Type"`               // TAG_Int(4) = 0
}

// ID ...
func (*Banner) ID() string {
	return IDBanner
}

func (b *Banner) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
	protocol.NBTInt(&b.Base, io.Varuint32)
	io.Varint32(&b.Type)
	protocol.NBTSliceVarint16Length(io, &b.Patterns, &[]fields.BannerPatterns{})
}
