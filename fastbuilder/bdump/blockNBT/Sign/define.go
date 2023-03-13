package blockNBT_Sign

import blockNBT_global "phoenixbuilder/fastbuilder/bdump/blockNBT/Global"

// SignDatas 结构体用于描述告示牌的 NBT 在被解析后的数据
type SignDatas struct {
	TextOwner                   string // TextOwner(TAG_String) = ""
	IgnoreLighting              byte   // IgnoreLighting(TAG_Byte) = 0
	SignTextColor               int32  // SignTextColor(TAG_Int) = 0
	TextIgnoreLegacyBugResolved byte   // TextIgnoreLegacyBugResolved(TAG_Byte) = 0
	Text                        string // Text(TAG_String) = ""
}

// Sign 结构体用于描述一个完整的告示牌
type Sign struct {
	BlockEntityDatas *blockNBT_global.BlockEntityDatas // 该方块实体的详细数据
	SignDatas        SignDatas                         // 告示牌数据
}
