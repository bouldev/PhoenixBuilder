package GlobalAPI

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
)

// 用作铁砧的承重方块
const BlockUnderAnvil string = "glass"

// 描述各个维度的 ID
const (
	OverWorldID = byte(iota) // 主世界
	NetherID                 // 下界
	EndID                    // 末地
)

// 描述各个维度可放置方块的最高高度
const (
	OverWorld_MaxPosy = int32(319) // 主世界
	Nether_MaxPosy    = int32(127) // 下界
	End_MaxPosy       = int32(255) // 末地
)

// 描述各个维度可放置方块的最低高度
const (
	OverWorld_MinPosy = int32(-64) // 主世界
	Nether_MinPosy    = int32(0)   // 下界
	End_MinPosy                    // 末地
)

// 描述一个空气物品
var AirItem protocol.ItemInstance = protocol.ItemInstance{
	StackNetworkID: 0,
	Stack: protocol.ItemStack{
		ItemType: protocol.ItemType{
			NetworkID:     0,
			MetadataValue: 0,
		},
		BlockRuntimeID: 0,
		Count:          0,
		NBTData:        map[string]interface{}(nil),
		CanBePlacedOn:  []string(nil),
		CanBreak:       []string(nil),
		HasNetworkID:   false,
	},
}

// 用于关闭容器时却发现到容器从未被打开时的报错信息
var ErrContainerNerverOpened error = fmt.Errorf(
	"CloseContainer: Container have been nerver opened",
)
