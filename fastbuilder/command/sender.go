package command

import (
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"github.com/google/uuid"
	"sync"
)

var UUIDMap sync.Map//= make(map[string]func(*minecraft.Conn,*[]protocol.CommandOutputMessage))
var BlockUpdateSubscribeMap sync.Map

func ClearUUIDMap() {
	UUIDMap=sync.Map{}
}

func SendCommand(command string, UUID uuid.UUID, conn *minecraft.Conn) error {
	requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	origin := protocol.CommandOrigin{
		Origin:         protocol.CommandOriginPlayer,
		UUID:           UUID,
		RequestID:      requestId.String(),
		PlayerUniqueID: 0,
	}
	commandRequest := &packet.CommandRequest{
		CommandLine:   command,
		CommandOrigin: origin,
		Internal:      false,
		UnLimited:     false,
	}
	return conn.WritePacket(commandRequest)
}

func SendWSCommand(command string, UUID uuid.UUID, conn *minecraft.Conn) error {
	requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	origin := protocol.CommandOrigin{
		Origin:         protocol.CommandOriginAutomationPlayer,
		UUID:           UUID,
		RequestID:      requestId.String(),
		PlayerUniqueID: 0,
	}
	commandRequest := &packet.CommandRequest{
		CommandLine:   command,
		CommandOrigin: origin,
		Internal:      false,
		UnLimited:     false,
	}
	return conn.WritePacket(commandRequest)
}

func SendSizukanaCommand(command string, conn *minecraft.Conn) error {
	return conn.WritePacket(&packet.SettingsCommand{
		CommandLine: command,
		SuppressOutput: true,
	})
}

func SendChat(content string, conn *minecraft.Conn) error {
	idd:=conn.IdentityData()
	return conn.WritePacket(&packet.Text {
		TextType: packet.TextTypeChat,
		NeedsTranslation: false,
		SourceName: idd.DisplayName,
		Message: content,
		XUID: idd.XUID,
	})
}