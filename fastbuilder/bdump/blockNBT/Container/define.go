package blockNBT_Container

import (
	"fmt"
	blockNBT_global "phoenixbuilder/fastbuilder/bdump/blockNBT/Global"
)

// Item 结构体用于描述单个的物品
type Item struct {
	Name   string // Name(TAG_String) = ""
	Count  uint8  // Count(TAG_Byte) = 0
	Damage uint16 // TAG_Short = 0
	Slot   uint8  // Slot(TAG_Byte) = 0
}

// Container 结构体用于描述一个容器
type Container struct {
	BlockEntityDatas *blockNBT_global.BlockEntityDatas // 该方块实体的详细数据
	Items            []Item                            // 容器内的物品数据
}

/*
此表描述了可被 replaceitem 生效的容器

key 代表容器的方块名，而 value 则代表此容器放置物品所使用的复合标签或列表
*/
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

// 用于 decode.go 中的 ReplaceNBTMapToContainerList 等函数
var KeyName string = "datas"
