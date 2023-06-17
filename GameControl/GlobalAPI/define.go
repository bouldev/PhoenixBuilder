package GlobalAPI

import (
	"phoenixbuilder/GameControl/ResourcesControlCenter"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 描述客户端的基本信息
type BotInfo struct {
	BotName      string // 客户端的游戏昵称
	BotIdentity  string // 客户端的唯一标识符 [当前还未使用]
	BotUniqueID  int64  // 客户端的唯一 ID [当前还未使用]
	BotRunTimeID uint64 // 客户端的运行时 ID
}

// 用于 PhoenixBuilder 与租赁服交互。
// 此结构体下的实现将允许您与租赁服进行交互操作，例如打开容器等
type GlobalAPI struct {
	// 用于向租赁服发送数据包的函数
	WritePacket func(packet.Packet) error
	// 存储客户端的基本信息
	BotInfo BotInfo
	// PhoenixBuilder 的各类公用资源
	Resources *ResourcesControlCenter.Resources
}
