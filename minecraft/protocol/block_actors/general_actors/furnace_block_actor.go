package general

import (
	"phoenixbuilder/minecraft/protocol"
)

// 描述 熔炉、高炉、烟熏炉 的通用字段
type FurnaceBlockActor struct {
	BlockActor   `mapstructure:",squash"`
	BurnDuration int16 `mapstructure:"BurnDuration"` // TAG_Short(3) = 0
	BurnTime     int16 `mapstructure:"BurnTime"`     // TAG_Short(3) = 0
	CookTime     int16 `mapstructure:"CookTime"`     // TAG_Short(3) = 0
	Items        []any `mapstructure:"Items"`        // TAG_List[TAG_Compound] (9[10])
	StoredXPInt  int32 `mapstructure:"StoredXPInt"`  // TAG_Int(4) = 0
}

func (f *FurnaceBlockActor) Marshal(r protocol.IO) {
	protocol.Single(r, &f.BlockActor)
	r.Varint16(&f.BurnTime)
	r.Varint16(&f.CookTime)
	r.Varint16(&f.BurnDuration)
	r.Varint32(&f.StoredXPInt)
	protocol.NBTSlice(r, &f.Items, func(t *[]protocol.ItemWithSlot) { r.ItemList(t) })
}
