package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 信标
type Beacon struct {
	general.BlockActor `mapstructure:",squash"`
	Primary            int32 `mapstructure:"primary"`   // TAG_Int(4) = 0
	Secondary          int32 `mapstructure:"secondary"` // TAG_Int(4) = 0
}

// ID ...
func (*Beacon) ID() string {
	return IDBeacon
}

func (b *Beacon) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
	io.Varint32(&b.Primary)
	io.Varint32(&b.Secondary)
}
