package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 潮涌核心
type Conduit struct {
	general.BlockActor `mapstructure:",squash"`
	Active             byte  `mapstructure:"Active"` // TAG_Byte(1) = 0
	Target             int64 `mapstructure:"Target"` // TAG_Long(5) = -1
}

// ID ...
func (*Conduit) ID() string {
	return IDConduit
}

func (c *Conduit) Marshal(io protocol.IO) {
	protocol.Single(io, &c.BlockActor)
	io.Varint64(&c.Target)
	io.Uint8(&c.Active)
}
