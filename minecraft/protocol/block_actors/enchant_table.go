package block_actors

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
	general "phoenixbuilder/minecraft/protocol/block_actors/general_actors"
)

// 附魔台
type EnchantTable struct {
	general.BlockActor
	Rotation float32 `nbt:"rott"`       // TAG_Float(6) = 0
	Name     string  `nbt:"CustomName"` // TAG_String(8) = ""
}

// ID ...
func (*EnchantTable) ID() string {
	return IDEnchantTable
}

func (e *EnchantTable) Marshal(io protocol.IO) {
	protocol.Single(io, &e.BlockActor)
	io.String(&e.Name)
	io.Float32(&e.Rotation)
}

func (e *EnchantTable) ToNBT() map[string]any {
	if len(e.Name) > 0 {
		temp := e.CustomName
		defer func() {
			e.CustomName = temp
		}()
		e.CustomName = e.Name
	}
	return utils.MergeMaps(
		e.BlockActor.ToNBT(),
		map[string]any{
			"rott": e.Rotation,
		},
	)
}

func (e *EnchantTable) FromNBT(x map[string]any) {
	e.BlockActor.FromNBT(x)
	e.Rotation = x["rott"].(float32)
}
