package GameInterface

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/fastbuilder/args"
	"phoenixbuilder/fastbuilder/py_rpc"
	cts "phoenixbuilder/fastbuilder/py_rpc/mod_event/client_to_server"
	cts_mc "phoenixbuilder/fastbuilder/py_rpc/mod_event/client_to_server/minecraft"
	cts_mc_a "phoenixbuilder/fastbuilder/py_rpc/mod_event/client_to_server/minecraft/ai_command"
	mei "phoenixbuilder/fastbuilder/py_rpc/mod_event/interface"
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
	if args.SkipMCPCheckChallenges {
		err := g.SendAICommand(command)
		if err != nil {
			return fmt.Errorf("SendSettingsCommand: %v", err)
		}
		return nil
	}
	// for restrictive situation
	if dimensional {
		command = fmt.Sprintf(
			`execute @a[name="%s"] ~ ~ ~ %s`,
			g.ClientInfo.DisplayName,
			command,
		)
	}
	// construct commands
	err := g.WritePacket(&packet.SettingsCommand{
		CommandLine:    command,
		SuppressOutput: true,
	})
	if err != nil {
		return fmt.Errorf("SendSettingsCommand: %v", err)
	}
	// send packet
	return nil
	// return
}

// 以 origin 的身份向租赁服发送命令且无视返回值。
// 属于私有实现
func (g *GameInterface) send_command(
	command string,
	uniqueId uuid.UUID,
	origin uint32,
) error {
	if args.SkipMCPCheckChallenges {
		err := g.send_netease_ai_command(command, uniqueId)
		if err != nil {
			return fmt.Errorf("send_command: %v", err)
		}
		return nil
	}
	// for restrictive situation
	pkt := packet.CommandRequest{
		CommandLine: command,
		CommandOrigin: protocol.CommandOrigin{
			Origin: origin,
			UUID:   uniqueId,
		},
		Internal:  false,
		UnLimited: false,
	}
	if origin == protocol.CommandOriginAutomationPlayer {
		pkt.CommandOrigin.RequestID = DefaultCommandRequestID
	}
	// construct command request packet
	err := g.WritePacket(&pkt)
	if err != nil {
		return fmt.Errorf("send_command: %v", err)
	}
	// send packet
	return nil
	// return
}

// 向租赁服发送魔法指令且无视返回值。
// 单独调用此函数会造成程序 panic 。
// 属于私有实现
func (g *GameInterface) send_netease_ai_command(
	command string,
	uniqueId uuid.UUID,
) error {
	event := cts_mc_a.ExecuteCommandEvent{
		CommandLine:      command,
		CommandRequestID: uniqueId,
	}
	module := cts_mc.AICommand{Module: &mei.DefaultModule{Event: &event}}
	park := cts.Minecraft{Default: mei.Default{Module: &module}}
	// construct request
	err := g.WritePacket(&packet.PyRpc{
		Value: py_rpc.Marshal(&py_rpc.ModEvent{
			Package: &park,
			Type:    py_rpc.ModEventClientToServer,
		}),
	})
	if err != nil {
		return fmt.Errorf("send_netease_ai_command: %v", err)
	}
	// send packet
	return nil
	// return
}

/*
以 origin 的身份向租赁服发送命令。
当 origin 不存在时，将发送魔法指令。

options 用于指定当次命令请求的自定义设置项。
如果要求不跟踪命令请求的响应体，
则数据包成功发送后将会返回空值。

属于私有实现
*/
func (g *GameInterface) send_command_with_options(
	command string,
	options ResourcesControl.CommandRequestOptions,
	origin *uint32,
) *ResourcesControl.CommandRespond {
	var err error
	command_request_id := ResourcesControl.GenerateUUID()
	// 初始化
	{
		var command_type string
		// 初始化
		switch origin {
		case nil:
			command_type = ResourcesControl.CommandTypeAICommand
		default:
			command_type = ResourcesControl.CommandTypeStandard
		}
		// 取得命令类型
		if args.SkipMCPCheckChallenges {
			command_type = ResourcesControl.CommandTypeAICommand
		}
		// 类型微调
		err = g.Resources.Command.WriteRequest(command_request_id, options, command_type)
		if err != nil {
			return &ResourcesControl.CommandRespond{
				Error:     fmt.Errorf("send_command_with_options: %v", err),
				ErrorType: ResourcesControl.ErrCommandRequestOthers,
			}
		}
		// 写入请求
	}
	// 写入请求到等待队列
	switch origin {
	case nil:
		err = g.send_netease_ai_command(command, command_request_id)
	default:
		err = g.send_command(command, command_request_id, *origin)
	}
	if err != nil {
		return &ResourcesControl.CommandRespond{
			Error:     fmt.Errorf("send_command_with_options: %v", err),
			ErrorType: ResourcesControl.ErrCommandRequestOthers,
		}
	}
	// 发送命令
	if options.WithNoResponse {
		return nil
	}
	// 如果不需要跟踪响应体
	resp := g.Resources.Command.LoadResponseAndDelete(command_request_id)
	if resp.Error != nil {
		resp.Error = fmt.Errorf("send_command_with_options: %v", resp.Error)
	}
	// 加载响应体
	if args.SkipMCPCheckChallenges && origin != nil {
		resp.Type = ResourcesControl.CommandTypeStandard
		// 覆写原始命令请求的类型
		switch resp.Respond {
		case nil:
			fake_resp := DefaultCommandOutput
			fake_resp.CommandOrigin.Origin = *origin
			fake_resp.CommandOrigin.UUID = command_request_id
			fake_resp.OutputMessages = []protocol.CommandOutputMessage{
				{
					Success:    false,
					Message:    "commands.generic.syntax",
					Parameters: []string{"", command, ""},
				},
			}
			if *origin == protocol.CommandOriginAutomationPlayer {
				fake_resp.DataSet = "{\n   \"statusCode\" : -2147483648\n}\n"
			}
			resp.Respond = &fake_resp
		default:
			resp.Respond.CommandOrigin.Origin = *origin
		}
		// 覆写 命令来源 和 命令输出 相关的字段
		switch *origin {
		case protocol.CommandOriginAutomationPlayer:
			resp.Respond.CommandOrigin.RequestID = DefaultCommandRequestID
			resp.Respond.OutputType = packet.CommandOutputTypeDataSet
		default:
			resp.Respond.CommandOrigin.RequestID = ""
			resp.Respond.OutputType = packet.CommandOutputTypeNone
			resp.Respond.DataSet = ""
		}
		// 覆写 请求ID、输出类型 和 数据集
	}
	// 针对限制性情况的响应体微调
	return &resp
	// 返回值
}

// 以玩家的身份向租赁服发送命令且无视返回值
func (g *GameInterface) SendCommand(command string) error {
	if args.SkipMCPCheckChallenges {
		err := g.SendAICommand(command)
		if err != nil {
			return fmt.Errorf("SendCommand: %v", err)
		}
		return nil
	}
	// for restrictive situation
	uniqueId, _ := uuid.NewUUID()
	err := g.send_command(command, uniqueId, protocol.CommandOriginPlayer)
	if err != nil {
		return fmt.Errorf("SendCommand: %v", err)
	}
	return nil
	// for normal situation
}

// 向租赁服发送 WS 命令且无视返回值
func (g *GameInterface) SendWSCommand(command string) error {
	if args.SkipMCPCheckChallenges {
		err := g.SendAICommand(command)
		if err != nil {
			return fmt.Errorf("SendWSCommand: %v", err)
		}
		return nil
	}
	// for restrictive situation
	uniqueId, _ := uuid.NewUUID()
	err := g.send_command(command, uniqueId, protocol.CommandOriginAutomationPlayer)
	if err != nil {
		return fmt.Errorf("SendWSCommand: %v", err)
	}
	return nil
	// for normal situation
}

// 向租赁服发送 魔法指令 且无视返回值
func (g *GameInterface) SendAICommand(command string) error {
	holder := g.Resources.Command.Occupy()
	defer g.Resources.Command.Release(holder)
	uniqueId, _ := uuid.NewUUID()
	// prepare
	err := g.Resources.Command.WriteRequest(
		uniqueId,
		ResourcesControl.CommandRequestOptions{
			WithNoResponse: true,
		},
		ResourcesControl.CommandTypeAICommand,
	)
	if err != nil {
		return fmt.Errorf("SendAICommand: %v", err)
	}
	// write request
	err = g.send_netease_ai_command(command, uniqueId)
	if err != nil {
		return fmt.Errorf("SendAICommand: %v", err)
	}
	// send command
	return nil
	// return
}

// 以玩家的身份向租赁服发送命令且获取返回值。
// options 指定当次命令请求的自定义设置项
func (g *GameInterface) SendCommandWithResponse(
	command string,
	options ResourcesControl.CommandRequestOptions,
) ResourcesControl.CommandRespond {
	origin := uint32(protocol.CommandOriginPlayer)
	resp := g.send_command_with_options(command, options, &origin)
	if resp.Error != nil {
		resp.Error = fmt.Errorf("SendCommandWithResponse: %v", resp.Error)
	}
	return *resp
}

// 向租赁服发送 WS 命令且获取返回值。
// options 指定当次命令请求的自定义设置项
func (g *GameInterface) SendWSCommandWithResponse(
	command string,
	options ResourcesControl.CommandRequestOptions,
) ResourcesControl.CommandRespond {
	origin := uint32(protocol.CommandOriginAutomationPlayer)
	resp := g.send_command_with_options(command, options, &origin)
	if resp.Error != nil {
		resp.Error = fmt.Errorf("SendWSCommandWithResponse: %v", resp.Error)
	}
	return *resp
}

// 向租赁服发送 魔法指令 且获取返回值。
// options 指定当次命令请求的自定义设置项
func (g *GameInterface) SendAICommandWithResponse(
	command string,
	options ResourcesControl.CommandRequestOptions,
) ResourcesControl.CommandRespond {
	resp := g.send_command_with_options(command, options, nil)
	if resp.Error != nil {
		resp.Error = fmt.Errorf("SendAICommandWithResponse: %v", resp.Error)
	}
	return *resp
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
