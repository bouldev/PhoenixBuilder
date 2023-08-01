package GameInterface

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/args"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"
	"time"

	"github.com/google/uuid"
)

// 向租赁服发送 Sizukana 命令且无视返回值。
// 当 dimensional 为真时，
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
// options 指定当次命令请求的自定义设置项。
// 属于私有实现
func (g *GameInterface) sendCommandWithResponse(
	command string,
	options ResourcesControl.CommandRequestOptions,
	origin uint32,
) ResourcesControl.CommandRespond {
	uniqueId := ResourcesControl.GenerateUUID()
	err := g.Resources.Command.WriteRequest(uniqueId, options)
	if err != nil {
		return ResourcesControl.CommandRespond{
			Error:     fmt.Errorf("sendCommandWithResponse: %v", err),
			ErrorType: ResourcesControl.ErrCommandRequestOthers,
		}
	}
	// 写入请求到等待队列
	err = g.sendCommand(command, uniqueId, origin)
	if err != nil {
		return ResourcesControl.CommandRespond{
			Error:     fmt.Errorf("sendCommandWithResponse: %v", err),
			ErrorType: ResourcesControl.ErrCommandRequestOthers,
		}
	}
	// 发送命令
	resp := g.Resources.Command.LoadResponseAndDelete(uniqueId)
	if resp.Error != nil {
		resp.Error = fmt.Errorf("sendCommandWithResponse: %v", resp.Error)
	}
	return resp
	// 获取响应体并返回值
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

// 以玩家的身份向租赁服发送命令且获取返回值。
// options 指定当次命令请求的自定义设置项
func (g *GameInterface) SendCommandWithResponse(
	command string,
	options ResourcesControl.CommandRequestOptions,
) ResourcesControl.CommandRespond {
	resp := g.sendCommandWithResponse(command, options, protocol.CommandOriginPlayer)
	if resp.Error != nil {
		resp.Error = fmt.Errorf("SendCommandWithResponse: %v", resp.Error)
	}
	return resp
}

// 向租赁服发送 WS 命令且获取返回值。
// options 指定当次命令请求的自定义设置项
func (g *GameInterface) SendWSCommandWithResponse(
	command string,
	options ResourcesControl.CommandRequestOptions,
) ResourcesControl.CommandRespond {
	resp := g.sendCommandWithResponse(command, options, protocol.CommandOriginAutomationPlayer)
	if resp.Error != nil {
		resp.Error = fmt.Errorf("SendWSCommandWithResponse: %v", resp.Error)
	}
	return resp
}

/*
一个基于游戏刻流逝的通用实现，
用于等待租赁服完成更改。

该实现效率较低，正常状态下，
平均每次调用将会阻滞 4~6 个游戏刻。

该实现也并不完全安全，
在 TPS 较低的情况下，实际表现欠佳。

遗憾地，由于现在没有良好的办法用于检测更改，
该方法因此被广泛使用且难以被替代
*/
func (g *GameInterface) AwaitChangesGeneral() error {
	olderTick, err := g.GetCurrentTick()
	if err != nil {
		return fmt.Errorf("AwaitChangesGeneral: %v", err)
	}
	// 获取当前的游戏刻
	timer := time.NewTicker(time.Second / 100)
	defer timer.Stop()
	// 设定计时器
	for {
		newerTick, err := g.GetCurrentTick()
		if err != nil {
			return fmt.Errorf("AwaitChangesGeneral: %v", err)
		}
		if newerTick >= olderTick+2 {
			return nil
		}
		<-timer.C
	}
	// 等待游戏刻流逝至少 2 游戏刻
}

// 将 content 打印到终端，
// 必要时还会以聊天的形式输出到游戏中
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
