package block_actors

import (
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 磁石
type Lodestone struct {
	general.BlockActor `mapstructure:",squash"`
	TrackingHandle     *int32 `mapstructure:"trackingHandle,omitempty"` // TAG_Int(4) = 0
}

// ID ...
func (*Lodestone) ID() string {
	return IDLodestone
}

func (l *Lodestone) Marshal(io protocol.IO) {
	f := func() *int32 {
		if l.TrackingHandle == nil {
			l.TrackingHandle = new(int32)
		}
		return l.TrackingHandle
	}

	protocol.Single(io, &l.BlockActor)
	protocol.NBTOptionalFunc(io, l.TrackingHandle, f, true, io.Varint32)
}
