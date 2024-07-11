package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 钟
type Bell struct {
	general.BlockActor
	Direction int32 `nbt:"Direction"` // TAG_Int(4) = 255
	Ringing   byte  `nbt:"Ringing"`   // TAG_Byte(1) = 0
	Ticks     int32 `nbt:"Ticks"`     // TAG_Int(4) = 18
}

// ID ...
func (*Bell) ID() string {
	return IDBell
}

func (b *Bell) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
	io.Uint8(&b.Ringing)
	io.Varint32(&b.Ticks)
	io.Varint32(&b.Direction)
}

func (b *Bell) ToNBT() map[string]any {
	return utils.MergeMaps(
		b.BlockActor.ToNBT(),
		map[string]any{
			"Direction": b.Direction,
			"Ringing":   b.Ringing,
			"Ticks":     b.Ticks,
		},
	)
}

func (b *Bell) FromNBT(x map[string]any) {
	b.BlockActor.FromNBT(x)
	b.Direction = x["Direction"].(int32)
	b.Ringing = x["Ringing"].(byte)
	b.Ticks = x["Ticks"].(int32)
}
