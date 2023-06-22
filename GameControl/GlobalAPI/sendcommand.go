package GlobalAPI

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// 向租赁服发送 Sizukana 命令且无视返回值。
// 当 sendDimensionalCmd 为真时，
// 将使用 execute 更换命令执行环境为机器人所在的环境
func (g *GlobalAPI) SendSettingsCommand(
	command string,
	sendDimensionalCmd bool,
) error {
	if sendDimensionalCmd {
		command = fmt.Sprintf(
			`execute @a[name="%v"] ~ ~ ~ %v`,
			g.BotInfo.BotName,
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
func (g *GlobalAPI) sendCommandPrivate(
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
		return fmt.Errorf("sendCommandPrivate: %v", err)
	}
	return nil
}

// 以 origin 的身份向租赁服发送命令并且取得响应体。
// 属于私有实现
func (g *GlobalAPI) sendCMDWithRespPrivate(
	command string,
	origin uint32,
) (packet.CommandOutput, error) {
	uniqueId := generateUUID()
	err := g.Resources.Command.WriteRequest(uniqueId)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("sendCMDWithRespPrivate: %v", err)
	}
	// 写入请求到等待队列
	err = g.sendCommandPrivate(command, uniqueId, origin)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("sendCMDWithRespPrivate: %v", err)
	}
	// 发送命令
	ans, err := g.Resources.Command.LoadResponceAndDelete(uniqueId)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("sendCMDWithRespPrivate: %v", err)
	}
	// 等待租赁服响应命令请求并取得命令请求的返回值
	return ans, nil
	// 返回值
}

// 以玩家的身份向租赁服发送命令且无视返回值
func (g *GlobalAPI) SendCommand(command string, uniqueId uuid.UUID) error {
	err := g.sendCommandPrivate(command, uniqueId, protocol.CommandOriginPlayer)
	if err != nil {
		return fmt.Errorf("SendCommand: %v", err)
	}
	return nil
}

// 向租赁服发送 WS 命令且无视返回值
func (g *GlobalAPI) SendWSCommand(command string, uniqueId uuid.UUID) error {
	err := g.sendCommandPrivate(command, uniqueId, protocol.CommandOriginAutomationPlayer)
	if err != nil {
		return fmt.Errorf("SendWSCommand: %v", err)
	}
	return nil
}

// 以玩家的身份向租赁服发送命令且获取返回值
func (g *GlobalAPI) SendCommandWithResponce(command string) (packet.CommandOutput, error) {
	resp, err := g.sendCMDWithRespPrivate(command, protocol.CommandOriginPlayer)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("SendCommandWithResponce: %v", err)
	}
	return resp, nil
}

// 向租赁服发送 WS 命令且获取返回值
func (g *GlobalAPI) SendWSCommandWithResponce(command string) (packet.CommandOutput, error) {
	resp, err := g.sendCMDWithRespPrivate(command, protocol.CommandOriginAutomationPlayer)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponce: %v", err)
	}
	return resp, nil
}
