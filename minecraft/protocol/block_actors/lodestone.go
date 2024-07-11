package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 磁石
type Lodestone struct {
	general.BlockActor
	TrackingHandle protocol.Optional[int32] `nbt:"trackingHandle"` // TAG_Int(4) = 0
}

// ID ...
func (*Lodestone) ID() string {
	return IDLodestone
}

func (l *Lodestone) Marshal(io protocol.IO) {
	protocol.Single(io, &l.BlockActor)
	protocol.OptionalFunc(io, &l.TrackingHandle, io.Varint32)
}

func (l *Lodestone) ToNBT() map[string]any {
	var temp map[string]any
	if trackingHandle, has := l.TrackingHandle.Value(); has {
		temp = map[string]any{
			"trackingHandle": trackingHandle,
		}
	}
	return utils.MergeMaps(
		l.BlockActor.ToNBT(),
		temp,
	)
}

func (l *Lodestone) FromNBT(x map[string]any) {
	l.BlockActor.FromNBT(x)

	if trackingHandle, has := x["trackingHandle"].(int32); has {
		l.TrackingHandle = protocol.Optional[int32]{Set: true, Val: trackingHandle}
	}
}
