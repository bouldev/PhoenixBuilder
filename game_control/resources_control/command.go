package ResourcesControl

import (
	"fmt"
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
	c.response.Store(key, make(chan packet.CommandOutput, 1))
	return nil
	// return
}

// 尝试向请求 ID 为 key 的命令请求写入返回值 resp 。
// 属于私有实现。
// 如果 key 不存在，亦不会返回错误。
func (c *commandRequestWithResponse) tryToWriteResponse(
	key uuid.UUID,
	resp packet.CommandOutput,
) error {
	_, exist0 := c.request.Load(key)
	value, exist1 := c.response.Load(key)
	if !exist0 || !exist1 {
		return nil
	}
	// if key is not exist
	chanGet, normal := value.(chan packet.CommandOutput)
	if !normal {
		return fmt.Errorf("tryToWriteResponse: Failed to convert value into (chan packet.CommandOutput); value = %#v", value)
	}
	// convert data
	chanGet <- resp
	close(chanGet)
	return nil
	// return
}

// 读取请求 ID 为 key 的命令请求的响应体，
// 同时移除此命令请求
func (c *commandRequestWithResponse) LoadResponseAndDelete(key uuid.UUID) CommandRespond {
	options_origin, exist0 := c.request.Load(key)
	response_origin, exist1 := c.response.Load(key)
	if !exist0 || !exist1 {
		return CommandRespond{
			Error:     fmt.Errorf("LoadResponseAndDelete: %v is not recorded", key.String()),
			ErrorType: ErrCommandRequestNotRecord,
		}
	}
	// if key is not exist
	options_got, normal := options_origin.(CommandRequestOptions)
	if !normal {
		return CommandRespond{
			Error:     fmt.Errorf("LoadResponseAndDelete: Failed to convert options_origin into CommandRequestOptions; options_origin = %#v", options_origin),
			ErrorType: ErrCommandRequestConversionFailure,
		}
	}
	response_got, normal := response_origin.(chan packet.CommandOutput)
	if !normal {
		return CommandRespond{
			Error:     fmt.Errorf("LoadResponseAndDelete: Failed to convert response_origin into (chan packet.CommandOutput); response_origin = %#v", response_origin),
			ErrorType: ErrCommandRequestConversionFailure,
		}
	}
	// convert data
	{
		if options_got.TimeOut == CommandRequestNoDeadLine {
			res := <-response_got
			c.request.Delete(key)
			c.response.Delete(key)
			return CommandRespond{Respond: res}
		}
		// if there is no time limit
		select {
		case res := <-response_got:
			c.request.Delete(key)
			c.response.Delete(key)
			return CommandRespond{Respond: res}
		case <-time.After(options_got.TimeOut):
			c.request.Delete(key)
			c.response.Delete(key)
			return CommandRespond{
				Error:     fmt.Errorf(`LoadResponseAndDelete: Request "%v" time out`, key.String()),
				ErrorType: ErrCommandRequestTimeOut,
			}
		}
		// if there's a time limit
	}
	// process and return
}
