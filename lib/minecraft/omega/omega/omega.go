package omega

import (
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol/packet"
	"fastbuilder-core/lib/minecraft/mirror/define"
	"fastbuilder-core/lib/minecraft/omega/uq_holder"
	"github.com/google/uuid"
	"time"
)

// 与玩家交互的抽象
type PlayerKit interface {
	Say(msg string)
	RawSay(msg string)
	ActionBar(msg string)
	Title(msg string)
	SubTitle(msg string)
	GetRelatedUQ() *uq_holder.Player
	SetOnParamMsg(func(chat *GameChat) (catch bool)) error
	GetOnParamMsg() func(chat *GameChat) (catch bool)
	GetPos(selector string) chan *define.CubePos
}

// 一组发出指令的集合
type CmdSender interface {
	SendCmd(cmd string)
	SendCmdWithUUID(cmd string, ud uuid.UUID, ws bool)
	SendWOCmd(cmd string)
	SendCmdAndInvokeOnResponse(string, func(output *packet.CommandOutput))
	SendCmdAndInvokeOnResponseWithFeedback(string, func(output *packet.CommandOutput))
}

// 一组显示信息的集合
type InfoSender interface {
	BotSay(msg string)
	SayTo(target string, msg string)
	RawSayTo(target string, msg string)
	ActionBarTo(target string, msg string)
	TitleTo(target string, msg string)
	SubTitleTo(target string, msg string)
}

// 一组与玩家交互的集合
type PlayerInteract interface {
	GetPlayerKit(name string) PlayerKit
	GetPlayerKitByUUID(ud uuid.UUID) PlayerKit
	SetOnParamMsg(string, func(chat *GameChat) (catch bool)) error
}

// 表示可以放置一些特殊方块
type NBTBlockPlacer interface {
	PlaceCommandBlock(pos define.CubePos, commandBlockName string, blockDataOrStateStr string, withMove, withAirPrePlace bool, updatePacket *packet.CommandBlockUpdate,
		onDone func(done bool), timeOut time.Duration)
	PlaceSignBlock(pos define.CubePos, signBlockName string, blockDataOrStateStr string, withMove, withAirPrePlace bool, updatePacket *packet.BlockActorData, onDone func(done bool), timeOut time.Duration)
}

// 与游戏的交互接口，通过发出点什么来影响游戏
// 建议扩展该接口以提供更丰富的功能
type GameControlKernel interface {
	CmdSender
	InfoSender
	PlayerInteract
	NBTBlockPlacer
	SendMCPacket(packet.Packet)
}

// 表示框架可以对收到的游戏数据包做出响应，并进一步调用其子组件
type Reactable interface {
	React(packet.Packet)
}
