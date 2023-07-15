package omega

import (
	"phoenixbuilder/fastbuilder/lib/minecraft/mirror/define"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"

	"github.com/google/uuid"
)

// 可以向游戏发送数据包
type GameIntractable interface {
	SendPacket(packet.Packet)
	SendPacketBytes(pktID uint32, data []byte)
}

type GameReactable interface {
	HandlePacket(packet.Packet)
}

type PacketDispatcher interface {
	SetOnAnyPacketCallBack(func(packet.Packet))
	SetOnTypedPacketCallBack(uint32, func(packet.Packet))
}

type ReactCore interface {
	GameReactable
	PacketDispatcher
}

type InteractCore interface {
	GameIntractable
}

type CmdSender interface {
	SendWSCmd(cmd string)
	SendCmdWithUUID(cmd string, ud uuid.UUID, ws bool)
	SendWOCmd(cmd string)
	SendPlayerCmd(cmd string)
	SendWSCmdAndInvokeOnResponse(string, func(output *packet.CommandOutput))
	SendPlayerCmdAndInvokeOnResponseWithFeedback(string, func(output *packet.CommandOutput))
}

type InfoSender interface {
	BotSay(msg string)
	SayTo(target string, msg string)
	RawSayTo(target string, msg string)
	ActionBarTo(target string, msg string)
	TitleTo(target string, msg string)
	SubTitleTo(target string, msg string)
}

type NBTBlockPlacer interface {
	PlaceCommandBlock(pos define.CubePos, commandBlockName string, blockDataOrStateStr string, withMove, withAirPrePlace bool, updatePacket *packet.CommandBlockUpdate,
		onDone func(done bool), timeOut time.Duration)
	PlaceSignBlock(pos define.CubePos, signBlockName string, blockDataOrStateStr string, withMove, withAirPrePlace bool, updatePacket *packet.BlockActorData, onDone func(done bool), timeOut time.Duration)
}

type BlockPlacer interface {
	NBTBlockPlacer
}

type MicroOmega interface {
	GetGameControl() interface {
		InteractCore
		CmdSender
		InfoSender
		BlockPlacer
	}
	GetGameListener() PacketDispatcher
	GameReactable
	GetMicroUQHolder() MicroUQHolder
}

//type PlayerMsgListener interface {
//	SetOnParamMsg(playerName string, cb func(chat GameChat) (catch bool)) error
//}

// type PlayerInteract interface {
// 	GetPlayerKit(name string) PlayerKit
// 	GetPlayerKitByUUID(ud uuid.UUID) PlayerKit
// }
