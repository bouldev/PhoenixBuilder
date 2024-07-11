package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 下界反应核
type NetherReactor struct {
	general.BlockActor
	HasFinished   byte  `nbt:"HasFinished"`   // TAG_Byte(1) = 0
	IsInitialized byte  `nbt:"IsInitialized"` // TAG_Byte(1) = 0
	Progress      int16 `nbt:"Progress"`      // TAG_Short(3) = 0
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

func (n *NetherReactor) ToNBT() map[string]any {
	return utils.MergeMaps(
		n.BlockActor.ToNBT(),
		map[string]any{
			"HasFinished":   n.HasFinished,
			"IsInitialized": n.IsInitialized,
			"Progress":      n.Progress,
		},
	)
}

func (n *NetherReactor) FromNBT(x map[string]any) {
	n.BlockActor.FromNBT(x)
	n.HasFinished = x["HasFinished"].(byte)
	n.IsInitialized = x["IsInitialized"].(byte)
	n.Progress = x["Progress"].(int16)
}
