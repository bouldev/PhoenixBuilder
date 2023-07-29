package GameInterface

import (
	"fmt"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"
)

// ------------------------- general -------------------------

// 描述客户端的基本信息
type ClientInfo struct {
	// 机器人的游戏昵称
	DisplayName string
	// 客户端的唯一标识符 [当前还未使用]
	ClientIdentity string
	// 机器人登录到租赁服所使用的账户的 ID ，
	// 它在性质上应与 XBOX Live 的账户 ID 保持一致
	XUID string
	// 机器人的唯一 ID
	EntityUniqueID int64
	// 机器人的运行时 ID
	EntityRuntimeID uint64
}

// 用于 PhoenixBuilder 与租赁服交互。
// 此结构体下的实现将允许您与租赁服进行交互操作，例如打开容器等
type GameInterface struct {
	// 用于向租赁服发送数据包的函数
	WritePacket func(packet.Packet) error
	// 存储客户端的基本信息
	ClientInfo ClientInfo
	// PhoenixBuilder 的各类公用资源
	Resources *ResourcesControl.Resources
}

// ------------------------- constants -------------------------

// 描述最基本的四大选择器
const (
	TargetMySelf        string = "@s"
	TargetRandom        string = "@r"
	TargetAllEntity     string = "@e"
	TargetNearestPlayer string = "@p"
)

// 描述 Pick Block 请求的最长截止时间。
// 这与 packet.BlockPickRequest 相关。
// 当超过此时间后，将视为该请求未被接受
const (
	// 描述 Pick Block 请求的最长截止时间。
	// 这与 packet.BlockPickRequest 相关。
	// 当超过此时间后，将视为该请求未被接受
	BlockPickRequestDeadLine = time.Second
	// 在进行物品复制操作时等待容器数据到来的最长截止时间。
	// 当超过此时间后，将视为数据已经抵达
	ItemCopyWaitContainerDataComingDeadLine = time.Second
)

// 描述 Pick Block 失败后要重试的最大次数
const (
	// 容器打开失败后要重试的最大次数
	ContainerOpenReTryMaximumCounts = 3
	// 描述 Pick Block 失败后要重试的最大次数
	BlockPickRequestReTryMaximumCounts = 3
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

// 如果尝试移动空气到另外一个物品栏，则会返回该错误
var ErrMoveItemCheckFailure error = fmt.Errorf(
	"MoveItem: Item on source is air",
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
