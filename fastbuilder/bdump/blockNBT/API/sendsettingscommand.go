package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 向租赁服发送 Sizukana 命令且无视返回值
func (g *GlobalAPI) SendSettingsCommand(command string, sendDimensionalCmd bool) error {
	if sendDimensionalCmd {
		command = fmt.Sprintf(`execute @a[name="%v"] ~ ~ ~ %v`, g.BotName, command)
	}
	err := g.WritePacket(&packet.SettingsCommand{
		CommandLine:    command,
		SuppressOutput: true,
	})
	if err != nil {
		return fmt.Errorf("SendSettingsCommand: %v", err)
	}
	return nil
}
