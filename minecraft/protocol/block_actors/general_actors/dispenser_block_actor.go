package general

import (
	"phoenixbuilder/minecraft/protocol"
)

// 描述 发射器 和 投掷器 的通用字段
type DispenserBlockActor struct {
	RandomizableBlockActor `mapstructure:",squash"`
	Items                  []any `mapstructure:"Items"` // TAG_List[TAG_Compound] (9[10])
}

func (d *DispenserBlockActor) Marshal(r protocol.IO) {
	var name string = d.CustomName

	protocol.Single(r, &d.RandomizableBlockActor)
	protocol.NBTSlice(r, &d.Items, func(t *[]protocol.ItemWithSlot) { r.ItemList(t) })
	r.String(&name)

	if len(name) > 0 {
		d.CustomName = name
	}
}
