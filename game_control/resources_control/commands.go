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

/*
提交请求 ID 为 key 的命令请求。

request_type 指代原始的命令请求的类型，
例如这是一个 标准命令 ，亦或是一个 魔法指令。

options 指定当次命令请求的自定义设置项
*/
func (c *commandRequestWithResponse) WriteRequest(
	key uuid.UUID,
	options CommandRequestOptions,
	request_type string,
) error {
	c.request_lock.Lock()
	defer c.request_lock.Unlock()
	// prepare
	request := c.request.GetElement(key)
	_, exist := c.response.Load(key)
	if request != nil || exist {
		return fmt.Errorf("WriteRequest: %v has already existed", key.String())
	}
	// if key has already exist
	switch request_type {
	case CommandTypeStandard:
		c.response.Store(key, &CommandRespond{Type: request_type})
	case CommandTypeAICommand:
		c.response.Store(
			key, &CommandRespond{
				AICommand: &AICommandDetails{},
				Type:      request_type,
			},
		)
	default:
		return fmt.Errorf("WriteRequest: Unsupported request type %#v", request_type)
	}
	// set response
	c.request.Set(key, options)
	c.signal.Store(key, make(chan uint8, 2))
	// set request and init signal
	return nil
	// return
}

// 尝试向请求 ID 为 key 的命令请求写入返回值 resp 。
// 属于私有实现。
// 如果 key 不存在，亦不会返回错误
func (c *commandRequestWithResponse) try_to_write_response(
	key uuid.UUID,
	resp packet.CommandOutput,
) error {
	if len(resp.CommandOrigin.RequestID) == 0 {
		c.ai_command_resp = &resp
		return nil
	}
	// for netease ai command
	options := c.request.GetElement(key)
	channel, exist0 := c.signal.Load(key)
	response, exist1 := c.response.Load(key)
	if options == nil || !exist0 || !exist1 {
		return nil
	}
	// if key is not exist
	response.Respond = &resp
	channel <- SignalCouldLoadRespond
	// set data and send signal
	if options.Value.WithNoResponse {
		c.LoadResponseAndDelete(key)
	}
	// if we don't have to track the response
	return nil
	// return
}

// 处理 event 所指代的 魔法指令 的响应体。
// 如果响应体对应的命令请求未被找到，
// 则会造成程序 panic 。
// 属于私有实现
func (c *commandRequestWithResponse) on_ai_command(event stc_mc.AICommand) {
	defer func() {
		c.ai_command_resp = nil
	}()
	// clean up
	switch e := event.Module.(*mei.DefaultModule).Event.(type) {
	case *ai_command.ExecuteCommandOutputEvent:
		resp, exist0 := c.response.Load(e.CommandRequestID)
		channel, exist1 := c.signal.Load(e.CommandRequestID)
		if !exist0 || !exist1 {
			panic("on_ai_command: Attempt to send NeteaseAICommand(packet.PyRpc/CS2ModEvent/ExecuteCommandEvent) without using ResourcesControlCenter")
		}
		// load data by command request id
		if resp.Respond == nil && c.ai_command_resp != nil {
			resp.Respond = c.ai_command_resp
			channel <- SignalRespondReceived
		}
		// set standard response
		resp.AICommand.Output = append(resp.AICommand.Output, *e)
		// set output data
	case *ai_command.AvailableCheckFailed:
		c.request_lock.RLock()
		defer c.request_lock.RUnlock()
		var options CommandRequestOptions
		var command_request_id uuid.UUID
		var resp *CommandRespond
		// init
		for opt := c.request.Front(); opt != nil; opt = opt.Next() {
			value, exist := c.response.Load(opt.Key)
			if !exist {
				continue
			}
			if value.Type == CommandTypeAICommand {
				options = opt.Value
				command_request_id = opt.Key
				resp = value
				break
			}
		}
		// get the oldest ai command request
		channel, exist := c.signal.Load(command_request_id)
		if resp == nil || !exist {
			panic("on_ai_command: Attempt to send NeteaseAICommand(packet.PyRpc/CS2ModEvent/ExecuteCommandEvent) without using ResourcesControlCenter")
		}
		// load data and check
		resp.AICommand.Result.CommandRequestID = command_request_id
		resp.AICommand.PreCheckError = e
		channel <- SignalCouldLoadRespond
		// set data and send signal
		if options.WithNoResponse {
			c.LoadResponseAndDelete(command_request_id)
		}
		// if we don't have to track the response
	case *ai_command.AfterExecuteCommandEvent:
		c.request_lock.RLock()
		defer c.request_lock.RUnlock()
		// prepare
		options := c.request.GetElement(e.CommandRequestID)
		resp, exist0 := c.response.Load(e.CommandRequestID)
		channel, exist1 := c.signal.Load(e.CommandRequestID)
		if options == nil || !exist0 || !exist1 {
			panic("on_ai_command: Attempt to send NeteaseAICommand(packet.PyRpc/CS2ModEvent/ExecuteCommandEvent) without using ResourcesControlCenter")
		}
		// load data from command request id
		if resp.Respond == nil && c.ai_command_resp != nil {
			resp.Respond = c.ai_command_resp
			channel <- SignalRespondReceived
		}
		// set standard response
		resp.AICommand.Result = *e
		channel <- SignalCouldLoadRespond
		// set data and send signal
		if options.Value.WithNoResponse {
			c.LoadResponseAndDelete(e.CommandRequestID)
		}
		// if we don't have to track the response
	}
	// process each event of ai command
}

// 读取请求 ID 为 key 的命令请求的响应体，
// 同时移除此命令请求
func (c *commandRequestWithResponse) LoadResponseAndDelete(key uuid.UUID) CommandRespond {
	c.request_lock.RLock()
	options := c.request.GetElement(key)
	c.request_lock.RUnlock()
	channel, exist1 := c.signal.Load(key)
	// load data from key
	if options == nil || !exist1 {
		return CommandRespond{
			Error:     fmt.Errorf("LoadResponseAndDelete: %v is not recorded", key.String()),
			ErrorType: ErrCommandRequestNotRecord,
		}
	}
	// if key is not exist
	for {
		if options.Value.TimeOut == CommandRequestNoDeadLine {
			if flag := <-channel; flag != SignalCouldLoadRespond {
				continue
			}
			close(channel)
			c.request.Delete(key)
			c.signal.Delete(key)
			resp, _ := c.response.LoadAndDelete(key)
			return *resp
		}
		// if there is no time limit
		select {
		case flag := <-channel:
			if flag == SignalRespondReceived {
				options.Value.TimeOut = CommandRequestNoDeadLine
				continue
			}
			close(channel)
			c.request.Delete(key)
			c.signal.Delete(key)
			resp, _ := c.response.LoadAndDelete(key)
			return *resp
		case <-time.After(options.Value.TimeOut):
			c.request.Delete(key)
			c.signal.Delete(key)
			return CommandRespond{
				Error:     fmt.Errorf(`LoadResponseAndDelete: Request "%v" time out`, key.String()),
				ErrorType: ErrCommandRequestTimeOut,
			}
		}
		// if there's a time limit
	}
	// process and return
}
