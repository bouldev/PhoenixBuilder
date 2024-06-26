package GameInterface

import (
	"fmt"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"

	"github.com/google/uuid"
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

/*
一个约定俗成的常量，
其使用最早可被追溯到三年前，
且目前仍被广泛使用。

该字段的目的是未知的，
但我们至今仍然保留这一传统，
因为它可以有助于区分命令响
应体对应的原始请求的类型。

该字段被用于指代
packet.CommandRequest 中
的 CommandOrigin.RequestID 字段，
且应当仅被用于 Websocket 命令的发送
*/
const DefaultCommandRequestID = "96045347-a6a3-4114-94c0-1bc4cc561694"

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

const (
	// 容器 打开/关闭 失败后要重试的最大次数
	ContainerOperationsReTryMaximumCounts = 3
	// 描述 Pick Block 失败后要重试的最大次数
	BlockPickRequestReTryMaximumCounts = 3
)

// 用作铁砧的承重方块
const (
	// 用作铁砧的承重方块
	AnvilBase string = "minecraft:glass"
	/*
		用作放置方块时的依赖性方块。

		部分方块需要客户端以点击方块的形式来放置，
		例如告示牌和不同朝向的潜影盒。
		这里则选择了绿宝石块作为被点击的方块。

		SuperScript 最喜欢绿宝石块了！
	*/
	PlaceBlockBase string = "minecraft:emerald_block"
)

// 描述各个容器的容器 ID
const (
	// chest, trapped_chest,
	// hopper, dispenser,
	// dropper
	ContainerIDDefault = byte(protocol.ContainerLevelEntity) // 7

	ContainerIDInventory = byte(protocol.ContainerCombinedHotBarAndInventory) // 12
	ContainerIDHotBar    = byte(protocol.ContainerHotBar)                     // 29

	ContainerIDShulkerBox = byte(protocol.ContainerShulkerBox) // 31
	ContainerIDBarrel     = byte(protocol.ContainerBarrel)     // 59

	ContainerIDFurnaceIngredient = byte(protocol.ContainerFurnaceIngredient) // 26
	ContainerIDFurnaceFuel       = byte(protocol.ContainerFurnaceFuel)       // 25
	ContainerIDFurnaceResult     = byte(protocol.ContainerFurnaceResult)     // 27

	ContainerIDSmokerIngredient = byte(protocol.ContainerSmokerIngredient) // 47
	ContainerIDSmokerFuel       = ContainerIDFurnaceFuel                   // 25
	ContainerIDSmokerResult     = ContainerIDFurnaceResult                 // 27

	ContainerIDBlastFurnaceIngredient = byte(protocol.ContainerBlastFurnaceIngredient) // 46
	ContainerIDBlastFurnaceFuel       = ContainerIDFurnaceFuel                         // 25
	ContainerIDBlastFurnaceResult     = ContainerIDFurnaceResult                       // 27

	ContainerIDBrewingStandInput  = byte(protocol.ContainerBrewingStandInput)  // 9
	ContainerIDBrewingStandFuel   = byte(protocol.ContainerBrewingStandFuel)   // 11
	ContainerIDBrewingStandResult = byte(protocol.ContainerBrewingStandResult) // 10

	ContainerIDUnknown = byte(255)
	ContainerIDInvalid = ContainerIDUnknown // 255
)

// 描述各个维度的 ID
const (
	OverWorldID = byte(iota)
	NetherID
	EndID
)

// 描述各个维度可放置方块的最高高度
const (
	OverWorld_MaxPosy = int32(319)
	Nether_MaxPosy    = int32(127)
	End_MaxPosy       = int32(255)
)

// 描述各个维度可放置方块的最低高度
const (
	OverWorld_MinPosy = int32(-64)
	Nether_MinPosy    = int32(0)
	End_MinPosy
)

// 描述一个默认状态下的
// Websocket 命令请求所
// 对应的响应体
var DefaultCommandOutput = packet.CommandOutput{
	CommandOrigin: protocol.CommandOrigin{
		Origin:         protocol.CommandOriginAutomationPlayer,
		UUID:           uuid.UUID{},
		RequestID:      DefaultCommandRequestID,
		PlayerUniqueID: 0,
	},
	OutputType:     packet.CommandOutputTypeDataSet,
	SuccessCount:   0,
	OutputMessages: []protocol.CommandOutputMessage{},
	DataSet:        "",
}

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

// 用于容器操作相关的报错信息
var (
	// 用于打开容器时却发现到容器已被打开时的报错信息
	ErrContainerHasBeenOpened error = fmt.Errorf(
		"OpenContainer: Container has been opened",
	)
	// 用于关闭容器时却发现到容器从未被打开时的报错信息
	ErrContainerNerverOpened error = fmt.Errorf(
		"CloseContainer: Container have been nerver opened",
	)
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
