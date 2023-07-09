package GameInterface

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
)

// 用作铁砧的承重方块
const AnvilBase string = "glass"

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

/*
用于将字符串型的 uuid 通过下表的映射处理为新的字符串。
这么做是为了规避 NEMC 的屏蔽词问题。

相信这样处理过后的字符串不会再被屏蔽了，
或者，我们得说 *** 的 NEMC，你真好！
*/
var StringUUIDReplaceMap map[string]string = map[string]string{
	"0": "?", "1": "†", "2": "‡", "3": "⁎", "4": "⁕",
	"5": "⁑", "6": "⁜", "7": "⁂", "8": "✓", "9": "✕",
	"a": "⌁", "b": ",", "c": "_", "d": "~", "e": "!",
	"f": "@", "g": "#", "h": "♪", "i": "%", "j": "^",
	"k": "&", "l": "*", "m": "(", "n": ")", "o": "-",
	"p": "+", "q": "=", "r": "[", "s": "]", "t": "‰",
	"u": ";", "v": "'", "w": "⌀", "x": "<", "y": ">",
	"z": "‱",
}
