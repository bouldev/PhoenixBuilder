package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 床
type Bed struct {
	general.BlockActor
	Color uint32 `nbt:"color"` // * TAG_Byte(1) = 0
}

// ID ...
func (*Bed) ID() string {
	return IDBed
}

func (b *Bed) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
	io.Varuint32(&b.Color)
}

func (b *Bed) ToNBT() map[string]any {
	return utils.MergeMaps(
		b.BlockActor.ToNBT(),
		map[string]any{
			"color": byte(b.Color),
		},
	)
}

func (b *Bed) FromNBT(x map[string]any) {
	b.BlockActor.FromNBT(x)
	b.Color = uint32(x["color"].(byte))
}
