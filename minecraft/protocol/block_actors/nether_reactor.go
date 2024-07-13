package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 下界反应核
type NetherReactor struct {
	general.BlockActor `mapstructure:",squash"`
	HasFinished        byte  `mapstructure:"HasFinished"`   // TAG_Byte(1) = 0
	IsInitialized      byte  `mapstructure:"IsInitialized"` // TAG_Byte(1) = 0
	Progress           int16 `mapstructure:"Progress"`      // TAG_Short(3) = 0
}

// ID ...
func (*NetherReactor) ID() string {
	return IDNetherReactor
}

func (n *NetherReactor) Marshal(io protocol.IO) {
	protocol.Single(io, &n.BlockActor)
	io.Uint8(&n.IsInitialized)
	io.Varint16(&n.Progress)
	io.Uint8(&n.HasFinished)
}
