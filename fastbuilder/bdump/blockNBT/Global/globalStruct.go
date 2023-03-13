package blockNBT_global

import (
	blockNBT_API "phoenixbuilder/fastbuilder/bdump/blockNBT/API"
	"phoenixbuilder/fastbuilder/types"
)

// GeneralBlock 结构体用于描述通用型方块的数据
type GeneralBlock struct {
	Name   string                 // 方块名称(不含命名空间且应该全部小写)
	States map[string]interface{} // 方块状态
	NBT    map[string]interface{} // 当前方块所携带的 NBT 数据
}

// Datas 结构体用于描述一个方块实体的其他附加数据，例如方块的绝对坐标
type Datas struct {
	StatesString string            // 字符串形式的方块状态，用于在放置方块时使用
	Position     [3]int32          // 方块坐标(绝对坐标)
	Type         string            // 该方块的类型，例如各式各样的告示牌可以写作 Sign
	Settings     *types.MainConfig // 此参数应当只被 PhoenixBuilder 使用，除非 Omega 也需要设置一些功能
	FastMode     bool              // 是否是快速模式放置；若为真，则大多数方块实体的 NBT 数据将不会被 assign
	Others       interface{}       // 部分情况下可能会携带的不定数据，通常情况下应该为空
}

// BlockEntityDatas 是用于包装每个方块实体的结构体
type BlockEntityDatas struct {
	API   *blockNBT_API.GlobalAPI // 储存执行该方块状态放置所需的 API ，例如发包需要用到的函数等；此参数需要外部实现主动赋值，主要是为了兼容 Omega 和 PhoenixBuilder 对功能的同时使用
	Block GeneralBlock            // 一个通用型方块的数据，例如名称、方块状态和所携带的 NBT 数据
	Datas *Datas                  // 此方块的其他附加数据，例如方块的绝对坐标
}
