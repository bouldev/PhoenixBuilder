package blockNBT

import (
	"fmt"
	GlobalAPI "phoenixbuilder/GameControl/GlobalAPI"
	"phoenixbuilder/fastbuilder/types"
)

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

// Datas 结构体用于描述一个方块实体的其他附加数据，例如方块的绝对坐标
type Datas struct {
	// 字符串形式的方块状态，用于在放置方块时使用
	StatesString string
	// 方块坐标(绝对坐标)
	Position [3]int32
	// 该方块的类型，例如各式各样的告示牌可以写作 Sign
	Type string
	// 此参数应当只被 PhoenixBuilder 使用，除非 Omega 也需要设置一些功能
	Settings *types.MainConfig
	// 是否是快速模式放置；若为真，则大多数方块实体的 NBT 数据将不会被 assign
	FastMode bool
	// 部分情况下可能会携带的不定数据，通常情况下应该为空
	Others interface{}
}

// Package 是用于包装每个方块实体的结构体
type Package struct {
	// 储存执行该方块状态放置所需的 API ，例如发包需要用到的函数等
	// 此参数需要外部实现主动赋值，
	// 主要是为了兼容 Omega 和 PhoenixBuilder 对功能的同时使用
	API *GlobalAPI.GlobalAPI
	// 一个通用型方块的数据，例如名称、方块状态和所携带的 NBT 数据
	Block GeneralBlock
	// 此方块的其他附加数据，例如方块的绝对坐标
	Datas *Datas
}

// ------------------------- Container -------------------------

// 此表描述了可被 replaceitem 生效的容器。
// key 代表容器的方块名，而 value 则代表此容器放置物品所使用的复合标签或列表
var SupportContainerPool map[string]string = map[string]string{
	"blast_furnace":      "Items",
	"lit_blast_furnace":  "Items",
	"smoker":             "Items",
	"lit_smoker":         "Items",
	"furnace":            "Items",
	"lit_furnace":        "Items",
	"chest":              "Items",
	"barrel":             "Items",
	"trapped_chest":      "Items",
	"lectern":            "book",
	"hopper":             "Items",
	"dispenser":          "Items",
	"dropper":            "Items",
	"jukebox":            "RecordItem",
	"brewing_stand":      "Items",
	"undyed_shulker_box": "Items",
	"shulker_box":        "Items",
}

// 未被支持的容器会被用到以下两个变量
var NotASupportedContainer string = "Not a supported container"
var ErrNotASupportedContainer error = fmt.Errorf("replaceNBTMapToContainerList: Not a supported container")

// 用于 Container.go 中的 ReplaceNBTMapToContainerList 等函数
var KeyName string = "datas"

// ------------------------- END -------------------------
