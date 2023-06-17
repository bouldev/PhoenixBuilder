//go:build !is_tweak

package commands

import (
	"fmt"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
)

func (sender *CommandSender) GetBlockUpdateSubscribeMap() *sync.Map {
	return &sender.BlockUpdateSubscribeMap
}

func (sender *CommandSender) getConn() *minecraft.Conn {
	conn := sender.env.Connection.(*minecraft.Conn)
	return conn
}

func (sender *CommandSender) SendChat(content string) error {
	conn := sender.getConn()
	idd := conn.IdentityData()
	return conn.WritePacket(&packet.Text{
		TextType:         packet.TextTypeChat,
		NeedsTranslation: false,
		SourceName:       idd.DisplayName,
		Message:          content,
		XUID:             idd.XUID,
		PlayerRuntimeID:  fmt.Sprintf("%d", conn.GameData().EntityUniqueID),
	})
}
