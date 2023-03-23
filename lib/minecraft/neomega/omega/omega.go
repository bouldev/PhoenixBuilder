package omega

import (
	"fastbuilder-core/lib/minecraft/gophertunnel/protocol/packet"
	"github.com/google/uuid"
)

// 可以向游戏发送数据包
type GameIntractable interface {
	SendPacket(packet.Packet)
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
	SendCmd(cmd string)
	SendCmdWithUUID(cmd string, ud uuid.UUID, ws bool)
	SendWOCmd(cmd string)
	SendCmdAndInvokeOnResponse(string, func(output *packet.CommandOutput))
	SendCmdAndInvokeOnResponseWithFeedback(string, func(output *packet.CommandOutput))
}

type InfoSender interface {
	BotSay(msg string)
	SayTo(target string, msg string)
	RawSayTo(target string, msg string)
	ActionBarTo(target string, msg string)
	TitleTo(target string, msg string)
	SubTitleTo(target string, msg string)
}

type MicroOmega interface {
	GetGameControl() interface {
		InteractCore
		CmdSender
		InfoSender
	}
	GetGameListener() PacketDispatcher
	GameReactable
	GetBotInfo() BotBasicInfoHolder
}

//type PlayerMsgListener interface {
//	SetOnParamMsg(playerName string, cb func(chat GameChat) (catch bool)) error
//}

type PlayerInteract interface {
	GetPlayerKit(name string) PlayerKit
	GetPlayerKitByUUID(ud uuid.UUID) PlayerKit
}
