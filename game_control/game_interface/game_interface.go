package GameInterface

import (
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol/packet"
)

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
