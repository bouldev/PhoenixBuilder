package GameInterface

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"

	"github.com/google/uuid"
)

// 向租赁服发送 Sizukana 命令且无视返回值。
// 当 sendDimensionalCmd 为真时，
// 将使用 execute 更换命令执行环境为机器人所在的环境
func (g *GameInterface) SendSettingsCommand(
	command string,
	dimensional bool,
) error {
	if dimensional {
		command = fmt.Sprintf(
			`execute @a[name="%s"] ~ ~ ~ %s`,
			g.ClientInfo.DisplayName,
			command,
		)
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

// 以 origin 的身份向租赁服发送命令且无视返回值。
// 属于私有实现
func (g *GameInterface) sendCommand(
	command string,
	uniqueId uuid.UUID,
	origin uint32,
) error {
	requestId, _ := uuid.Parse("96045347-a6a3-4114-94c0-1bc4cc561694")
	err := g.WritePacket(&packet.CommandRequest{
		CommandLine: command,
		CommandOrigin: protocol.CommandOrigin{
			Origin:    origin,
			UUID:      uniqueId,
			RequestID: requestId.String(),
		},
		Internal:  false,
		UnLimited: false,
	})
	if err != nil {
		return fmt.Errorf("sendCommand: %v", err)
	}
	return nil
}

// 以 origin 的身份向租赁服发送命令并且取得响应体。
// 属于私有实现
func (g *GameInterface) sendCommandWithResponse(
	command string,
	origin uint32,
) (packet.CommandOutput, error) {
	uniqueId := generateUUID()
	err := g.Resources.Command.WriteRequest(uniqueId)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("sendCommandWithResponse: %v", err)
	}
	// 写入请求到等待队列
	err = g.sendCommand(command, uniqueId, origin)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("sendCommandWithResponse: %v", err)
	}
	// 发送命令
	res, err := g.Resources.Command.LoadResponseAndDelete(uniqueId)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("sendCommandWithResponse: %v", err)
	}
	// 等待租赁服响应命令请求并取得命令请求的返回值
	return res, nil
	// 返回值
}

// 以玩家的身份向租赁服发送命令且无视返回值
func (g *GameInterface) SendCommand(command string) error {
	uniqueId, _ := uuid.NewUUID()
	err := g.sendCommand(command, uniqueId, protocol.CommandOriginPlayer)
	if err != nil {
		return fmt.Errorf("SendCommand: %v", err)
	}
	return nil
}

// 向租赁服发送 WS 命令且无视返回值
func (g *GameInterface) SendWSCommand(command string) error {
	uniqueId, _ := uuid.NewUUID()
	err := g.sendCommand(command, uniqueId, protocol.CommandOriginAutomationPlayer)
	if err != nil {
		return fmt.Errorf("SendWSCommand: %v", err)
	}
	return nil
}

// 以玩家的身份向租赁服发送命令且获取返回值
func (g *GameInterface) SendCommandWithResponse(command string) (packet.CommandOutput, error) {
	resp, err := g.sendCommandWithResponse(command, protocol.CommandOriginPlayer)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("SendCommandWithResponse: %v", err)
	}
	return resp, nil
}

// 向租赁服发送 WS 命令且获取返回值
func (g *GameInterface) SendWSCommandWithResponse(command string) (packet.CommandOutput, error) {
	resp, err := g.sendCommandWithResponse(command, protocol.CommandOriginAutomationPlayer)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponse: %v", err)
	}
	return resp, nil
}

// ... [Need to add its use]
func (i *GameInterface) Output(content string) error {
	fmt.Printf("%s\n", content)
	if !args.InGameResponse {
		return nil
	}
	msg := strings.Replace(content, "schematic", "sc***atic", -1)
	msg = strings.Replace(msg, ".", "．", -1)
	return i.SendChat(fmt.Sprintf("§b%s", msg))
}

// 在聊天栏以聊天的形式输出消息 content
func (i *GameInterface) SendChat(content string) error {
	return i.WritePacket(&packet.Text{
		TextType:         packet.TextTypeChat,
		NeedsTranslation: false,
		SourceName:       i.ClientInfo.DisplayName,
		Message:          content,
		XUID:             i.ClientInfo.XUID,
		PlayerRuntimeID:  fmt.Sprintf("%d", i.ClientInfo.EntityUniqueID),
	})
}

// 以 actionbar 的形式向所有在线玩家显示 message
func (i *GameInterface) Title(message string) error {
	title_struct := map[string]interface{}{
		"rawtext": []interface{}{
			map[string]interface{}{
				"text": message,
			},
		},
	}
	json_content, _ := json.Marshal(title_struct)
	return i.SendSettingsCommand(fmt.Sprintf("titleraw @a actionbar %s", json_content), false)
}
