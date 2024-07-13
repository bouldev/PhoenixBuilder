package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 饰纹陶罐
type DecoratedPot struct {
	general.BlockActor `mapstructure:",squash"`
	Animation          byte          `mapstructure:"animation"` // Not used; TAG_Byte(1) = 0
	Item               protocol.Item `mapstructure:"item"`      // Not used; TAG_Compound(10)
}

// ID ...
func (*DecoratedPot) ID() string {
	return IDDecoratedPot
}

func (d *DecoratedPot) Marshal(io protocol.IO) {
	protocol.Single(io, &d.BlockActor)
}
