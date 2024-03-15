package ResourcesControl

import (
	"fmt"
	mei "phoenixbuilder/fastbuilder/py_rpc/py_rpc_content/mod_event/interface"
	stc_mc "phoenixbuilder/fastbuilder/py_rpc/py_rpc_content/mod_event/server_to_client/minecraft"
	"phoenixbuilder/fastbuilder/py_rpc/py_rpc_content/mod_event/server_to_client/minecraft/ai_command"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"

	"github.com/google/uuid"
)

// 提交请求 ID 为 key 的命令请求。
// options 指定当次命令请求的自定义设置项
func (c *commandRequestWithResponse) WriteRequest(
	key uuid.UUID,
	options CommandRequestOptions,
) error {
	_, exist0 := c.request.Load(key)
	_, exist1 := c.response.Load(key)
	if exist0 || exist1 {
		return fmt.Errorf("WriteRequest: %v has already existed", key.String())
	}
	// if key has already exist
	c.request.Store(key, options)
	c.couldLoadResp.Store(key, make(chan struct{}, 1))
	return nil
	// return
}

// 尝试向请求 ID 为 key 的命令请求写入返回值 resp 。
// 属于私有实现。
// 如果 key 不存在，亦不会返回错误
func (c *commandRequestWithResponse) tryToWriteResponse(
	key uuid.UUID,
	resp packet.CommandOutput,
) error {
	if len(resp.CommandOrigin.RequestID) == 0 {
		c.aiCommandResp = &resp
		return nil
	}
	// for netease ai command
	_, exist0 := c.request.Load(key)
	channel, exist1 := c.couldLoadResp.Load(key)
	if !exist0 || !exist1 {
		return nil
	}
	// if key is not exist
	c.response.Store(key, &CommandRespond{
		Respond:   &resp,
		AICommand: nil,
		Type:      CommandTypeStandard,
	})
	// set response
	channel <- struct{}{}
	return nil
	// send signal and return
}

// 处理 event 所指代的 魔法指令 的响应体。
// 属于私有实现。
// 如果对应的命令请求不存在，亦不会返回错误
func (c *commandRequestWithResponse) onAICommand(event stc_mc.AICommand) error {
	defer func() { c.aiCommandResp = nil }()
	// prepare
	switch e := event.Module.(*mei.DefaultModule).Event.(type) {
	case *ai_command.ExecuteCommandOutputEvent:
		resp, exist := c.response.Load(e.CommandRequestID)
		if !exist {
			return nil
		}
		resp.AICommand = &AICommandDetails{
			Output: e,
		}
	case *ai_command.AvailableCheckFailed:
		var command_request_id uuid.UUID
		var resp *CommandRespond
		// init
		c.response.Range(func(key uuid.UUID, value *CommandRespond) bool {
			if value.Type == CommandTypeAICommand {
				command_request_id = key
				resp = value
			}
			return true
		})
		// get one of ai command request
		channel, exist := c.couldLoadResp.Load(command_request_id)
		if resp == nil || !exist {
			return nil
		}
		// if channel not exist
		resp.AICommand.PreCheckError = e
		channel <- struct{}{}
		// set value and send signal
	case *ai_command.AfterExecuteCommandEvent:
		resp, exist0 := c.response.Load(e.CommandRequestID)
		channel, exist1 := c.couldLoadResp.Load(e.CommandRequestID)
		if !exist0 || !exist1 {
			return nil
		}
		resp.AICommand.Result = *e
		channel <- struct{}{}
	}
	// process each event of ai command
	return nil
	// return
}

// 读取请求 ID 为 key 的命令请求的响应体，
// 同时移除此命令请求
func (c *commandRequestWithResponse) LoadResponseAndDelete(key uuid.UUID) CommandRespond {
	options, exist0 := c.request.Load(key)
	channel, exist1 := c.couldLoadResp.Load(key)
	if !exist0 || !exist1 {
		return CommandRespond{
			Error:     fmt.Errorf("LoadResponseAndDelete: %v is not recorded", key.String()),
			ErrorType: ErrCommandRequestNotRecord,
		}
	}
	// if key is not exist
	{
		if options.TimeOut == CommandRequestNoDeadLine {
			<-channel
			close(channel)
			c.request.Delete(key)
			c.couldLoadResp.Delete(key)
			resp, _ := c.response.LoadAndDelete(key)
			return *resp
		}
		// if there is no time limit
		select {
		case <-channel:
			close(channel)
			c.request.Delete(key)
			c.couldLoadResp.Delete(key)
			resp, _ := c.response.LoadAndDelete(key)
			return *resp
		case <-time.After(options.TimeOut):
			c.request.Delete(key)
			c.couldLoadResp.Delete(key)
			return CommandRespond{
				Error:     fmt.Errorf(`LoadResponseAndDelete: Request "%v" time out`, key.String()),
				ErrorType: ErrCommandRequestTimeOut,
			}
		}
		// if there's a time limit
	}
	// process and return
}
