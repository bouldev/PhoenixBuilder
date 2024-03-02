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
	value <- resp
	close(value)
	return nil
	// return
}

// 读取请求 ID 为 key 的命令请求的响应体，
// 同时移除此命令请求
func (c *commandRequestWithResponse) LoadResponseAndDelete(key uuid.UUID) CommandRespond {
	options, exist0 := c.request.Load(key)
	response, exist1 := c.response.Load(key)
	if !exist0 || !exist1 {
		return CommandRespond{
			Error:     fmt.Errorf("LoadResponseAndDelete: %v is not recorded", key.String()),
			ErrorType: ErrCommandRequestNotRecord,
		}
	}
	// if key is not exist
	{
		if options.TimeOut == CommandRequestNoDeadLine {
			res := <-response
			c.request.Delete(key)
			c.response.Delete(key)
			return CommandRespond{Respond: res}
		}
		// if there is no time limit
		select {
		case res := <-response:
			c.request.Delete(key)
			c.response.Delete(key)
			return CommandRespond{Respond: res}
		case <-time.After(options.TimeOut):
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
