package NBTAssigner

import (
	"fmt"
	"phoenixbuilder/fastbuilder/environment/interfaces"
	"phoenixbuilder/fastbuilder/types"
)

// ------------------------- interface -------------------------

/*
GeneralBlockNBT 提供了一个通用的接口，
以便于您可以方便的解析对应的方块实体，
然后放置它并以最大的可能性注入 NBT 数据。

该接口实际与下方的 BlockEntity 结构体绑定
*/
type GeneralBlockNBT interface {
	// 解码
	Decode() error
	// 放置方块并尽可能注入 NBT
	WriteData() error
}

// ------------------------- general -------------------------

// GeneralBlock 结构体用于一个通用型的方块
type GeneralBlock struct {
	// 方块名称(不含命名空间且应该全部小写)
	Name string
	// 方块状态
	States map[string]interface{}
	// 当前方块所携带的 NBT 数据
	NBT map[string]interface{}
}

// AdditionalData 结构体用于描述一个方块实体的其他附加数据，例如方块的绝对坐标
type BlockAdditionalData struct {
	// 字符串形式的方块状态，用于在放置方块时使用
	BlockStates string
	// 方块坐标(绝对坐标)
	Position [3]int32
	// 该方块的类型，例如各式各样的告示牌可以写作 Sign
	// TODO: USE ENUM INSTEAD
	Type string
	// 此参数应当只被 PhoenixBuilder 使用，除非 Omega 也需要设置一些功能
	Settings *types.MainConfig
	// 是否是快速模式放置；若为真，则大多数方块实体的 NBT 数据将不会被注册
	FastMode bool
	// 部分情况下可能会携带的不定数据，通常情况下应该为空 [目前还未使用此字段]
	Others interface{}
}

// BlockEntity 是用于包装每个方块实体的结构体
type BlockEntity struct {
	// 储存放置该方块实体所需的 API ，例如发包需要用到的函数等
	// 此参数需要外部实现主动赋值，
	// 主要是为了兼容 Omega 和 PhoenixBuilder 对功能的同时使用
	Interface interfaces.GameInterface
	// 一个通用型方块的数据，例如名称、方块状态和所携带的 NBT 数据
	Block GeneralBlock
	// 此方块的其他附加数据，例如方块的绝对坐标
	AdditionalData BlockAdditionalData
}

// ------------------------- command_block -------------------------

// 描述单个命令方块中已解码的部分
type CommandBlockData struct {
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
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 存放已解码的命令方块数据
	CommandBlockData CommandBlockData
	// 为向下兼容而设，因为旧方法下不需要放置命令方块
	ShouldPlaceBlock bool
}

// ------------------------- container -------------------------

// 描述单个容器的固定数据，
// 这被用于 SupportContainerPool
type ContainerConstantData struct {
	// 描述此容器放置物品所使用的复合标签或列表
	StorageItemValue string
	// 描述此容器是否可以打开
	CouldOpen bool
	// 描述此容器的容器 ID 。
	// 为 255 时代表未被支持
	ContainerID uint8
}

// 描述一个容器
type Container struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	// 容器的内容物
	Contents []ItemPackage
}

// 未被支持的容器会被应用此错误信息。
// 用于 Container.go 中的 getContainerContents 等函数
var ErrNotASupportedContainer error = fmt.Errorf("getContainerContents: Not a supported container")

// ------------------------- sign -------------------------

// 描述一个告示牌
type Sign struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
}
