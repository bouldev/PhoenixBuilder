package general

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
)

// 描述 发射器 和 投掷器 的通用字段
type DispenserBlockActor struct {
	RandomizableBlockActor
	Items protocol.ItemList `nbt:"Items"`      // TAG_List[TAG_Compound] (9[10])
	Name  string            `nbt:"CustomName"` // TAG_String(8) = ""
}

func (d *DispenserBlockActor) Marshal(r protocol.IO) {
	protocol.Single(r, &d.RandomizableBlockActor)
	protocol.Single(r, &d.Items)
	r.String(&d.Name)
}

func (d *DispenserBlockActor) ToNBT() map[string]any {
	if len(d.Name) > 0 {
		temp := d.CustomName
		defer func() {
			d.CustomName = temp
		}()
		d.CustomName = d.Name
	}
	return utils.MergeMaps(
		d.RandomizableBlockActor.ToNBT(),
		map[string]any{
			"Items": d.Items.ToNBT(),
		},
	)
}

func (d *DispenserBlockActor) FromNBT(x map[string]any) {
	d.RandomizableBlockActor.FromNBT(x)
	d.Items.FromNBT(x["Items"].([]any))
}
