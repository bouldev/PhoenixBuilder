package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/block_actors/fields"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 化合物创建器
type ChemistryTable struct {
	general.BlockActor
	Item protocol.Optional[fields.ChemistryTableItem]
}

// ID ...
func (*ChemistryTable) ID() string {
	return IDChemistryTable
}

func (c *ChemistryTable) Marshal(io protocol.IO) {
	protocol.Single(io, &c.BlockActor)
	protocol.OptionalMarshaler(io, &c.Item)
}

func (c *ChemistryTable) ToNBT() map[string]any {
	var temp map[string]any
	if item, has := c.Item.Value(); has {
		temp = item.ToNBT()
	}
	return utils.MergeMaps(
		c.BlockActor.ToNBT(), temp,
	)
}

func (c *ChemistryTable) FromNBT(x map[string]any) {
	c.BlockActor.FromNBT(x)

	new := fields.ChemistryTableItem{}
	if new.CheckExist(x) {
		new.FromNBT(x)
		c.Item = protocol.Optional[fields.ChemistryTableItem]{Set: true, Val: new}
	}
}
