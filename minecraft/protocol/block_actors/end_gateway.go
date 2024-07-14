package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 末地折跃门方块
type EndGateway struct {
	general.BlockActor `mapstructure:",squash"`
	Age                int32   `mapstructure:"Age"`        // TAG_Int(4) = 0
	ExitPortal         []int32 `mapstructure:"ExitPortal"` // TAG_List[TAG_Int] (9[4])
}

// ID ...
func (*EndGateway) ID() string {
	return IDEndGateway
}

func (b *EndGateway) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
	io.Varint32(&b.Age)
	protocol.FuncSliceOfLen(io, 3, &b.ExitPortal, io.Varint32)
}
