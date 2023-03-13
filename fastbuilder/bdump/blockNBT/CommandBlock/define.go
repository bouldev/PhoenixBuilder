package blockNBT_CommandBlock

import blockNBT_global "phoenixbuilder/fastbuilder/bdump/blockNBT/Global"

// CommandBlockDatas 结构体用于描述命令方块的 NBT 在被解析后的数据
type CommandBlockDatas struct {
	Command            string // Command(TAG_String) = ""
	CustomName         string // CustomName(TAG_String) = ""
	LastOutput         string // LastOutput(TAG_String) = ""
	TickDelay          int32  // TickDelay(TAG_Int) = 0
	ExecuteOnFirstTick bool   // ExecuteOnFirstTick(TAG_Byte) = 1
	TrackOutput        bool   // TrackOutput(TAG_Byte) = 1
	ConditionalMode    bool   // conditionalMode(TAG_Byte) = 0
	Auto               bool   // auto(TAG_Byte) = 1
}

// CommandBlock 结构体用于描述一个完整的命令方块数据
type CommandBlock struct {
	BlockEntityDatas  *blockNBT_global.BlockEntityDatas // 该方块实体的详细数据
	CommandBlockDatas CommandBlockDatas                 // 命令方块数据
}
