package blockNBT

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
	Decode() error
	WriteData() error
}

// ------------------------- general -------------------------

// GeneralBlock 结构体用于描述通用型方块的数据
type GeneralBlock struct {
	// 方块名称(不含命名空间且应该全部小写)
	Name string
	// 方块状态
	States map[string]interface{}
	// 当前方块所携带的 NBT 数据
	NBT map[string]interface{}
}

// AdditionalData 结构体用于描述一个方块实体的其他附加数据，例如方块的绝对坐标
type AdditionalData struct {
	// 字符串形式的方块状态，用于在放置方块时使用
	BlockStates string
	// 方块坐标(绝对坐标)
	Position [3]int32
	// 该方块的类型，例如各式各样的告示牌可以写作 Sign
	// TODO: USE ENUM INSTEAD
	Type string
	// 此参数应当只被 PhoenixBuilder 使用，除非 Omega 也需要设置一些功能
	Settings *types.MainConfig
	// 是否是快速模式放置；若为真，则大多数方块实体的 NBT 数据将不会被 assign
	FastMode bool
	// 部分情况下可能会携带的不定数据，通常情况下应该为空
	Others interface{}
}

// BlockEntity 是用于包装每个方块实体的结构体
type BlockEntity struct {
	// 储存执行该方块状态放置所需的 API ，例如发包需要用到的函数等
	// 此参数需要外部实现主动赋值，
	// 主要是为了兼容 Omega 和 PhoenixBuilder 对功能的同时使用
	Interface interfaces.GameInterface
	// 一个通用型方块的数据，例如名称、方块状态和所携带的 NBT 数据
	Block GeneralBlock
	// 此方块的其他附加数据，例如方块的绝对坐标
	AdditionalData AdditionalData
}

// ------------------------- Container -------------------------

// 未被支持的容器会被用到以下两个变量
var ErrNotASupportedContainer error = fmt.Errorf("replaceNBTMapToContainerList: Not a supported container")

// 用于 Container.go 中的 ReplaceNBTMapToContainerList 等函数
var KeyName string = "datas"

// ------------------------- END -------------------------
