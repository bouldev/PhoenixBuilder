package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 比较器
type Comparator struct {
	general.BlockActor
	OutputSignal int32 `nbt:"OutputSignal"` // TAG_Int(4) = 0
}

// ID ...
func (*Comparator) ID() string {
	return IDComparator
}

func (c *Comparator) Marshal(io protocol.IO) {
	protocol.Single(io, &c.BlockActor)
	io.Varint32(&c.OutputSignal)
}

func (c *Comparator) ToNBT() map[string]any {
	return utils.MergeMaps(
		c.BlockActor.ToNBT(),
		map[string]any{
			"OutputSignal": c.OutputSignal,
		},
	)
}

func (c *Comparator) FromNBT(x map[string]any) {
	c.BlockActor.FromNBT(x)
	c.OutputSignal = x["OutputSignal"].(int32)
}
