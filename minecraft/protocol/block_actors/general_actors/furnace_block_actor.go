package general

import (
	"phoenixbuilder/fastbuilder/utils"
	"phoenixbuilder/minecraft/protocol"
)

// 描述 熔炉、高炉、烟熏炉 的通用字段
type FurnaceBlockActor struct {
	BlockActor
	BurnDuration int16             `nbt:"BurnDuration"` // TAG_Short(3) = 0
	BurnTime     int16             `nbt:"BurnTime"`     // TAG_Short(3) = 0
	CookTime     int16             `nbt:"CookTime"`     // TAG_Short(3) = 0
	Items        protocol.ItemList `nbt:"Items"`        // TAG_List[TAG_Compound] (9[10])
	StoredXPInt  int32             `nbt:"StoredXPInt"`  // TAG_Int(4) = 0
}

func (f *FurnaceBlockActor) Marshal(r protocol.IO) {
	protocol.Single(r, &f.BlockActor)
	r.Varint16(&f.BurnTime)
	r.Varint16(&f.CookTime)
	r.Varint16(&f.BurnDuration)
	r.Varint32(&f.StoredXPInt)
	protocol.Single(r, &f.Items)
}

func (f *FurnaceBlockActor) ToNBT() map[string]any {
	return utils.MergeMaps(
		f.BlockActor.ToNBT(),
		map[string]any{
			"BurnDuration": f.BurnDuration,
			"BurnTime":     f.BurnTime,
			"CookTime":     f.CookTime,
			"Items":        f.Items.ToNBT(),
			"StoredXPInt":  f.StoredXPInt,
		},
	)
}

func (f *FurnaceBlockActor) FromNBT(x map[string]any) {
	f.BlockActor.FromNBT(x)
	f.BurnDuration = x["BurnDuration"].(int16)
	f.BurnTime = x["BurnTime"].(int16)
	f.CookTime = x["CookTime"].(int16)
	f.Items.FromNBT(x["Items"].([]any))
	f.StoredXPInt = x["StoredXPInt"].(int32)
}
