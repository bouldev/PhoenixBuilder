package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// 向租赁服发送 WS 命令且无视返回值
func (g *GlobalAPI) SendWSCommand(command string, uniqueId uuid.UUID) error {
	requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	err := g.WritePacket(&packet.CommandRequest{
		CommandLine: command,
		CommandOrigin: protocol.CommandOrigin{
			Origin:    protocol.CommandOriginAutomationPlayer,
			UUID:      uniqueId,
			RequestID: requestId.String(),
		},
		Internal:  false,
		UnLimited: false,
	})
	if err != nil {
		return fmt.Errorf("SendWSCommand: %v", err)
	}
	return nil
}
