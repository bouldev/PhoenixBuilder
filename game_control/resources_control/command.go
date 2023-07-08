package ResourcesControl

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"

	"github.com/google/uuid"
)

// 提交请求 ID 为 key 的命令请求
func (c *commandRequestWithResponse) WriteRequest(key uuid.UUID) error {
	_, exist := c.requestWithResponse.Load(key)
	if exist {
		return fmt.Errorf("WriteRequest: %v has already existed", key.String())
	}
	// if key has already exist
	c.requestWithResponse.Store(key, make(chan packet.CommandOutput, 1))
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
	value, exist := c.requestWithResponse.Load(key)
	if !exist {
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

// 移除请求 ID 为 key 的命令请求，
// 主要被用于指令被网易屏蔽时的善后处理
func (c *commandRequestWithResponse) DeleteRequest(key uuid.UUID) {
	c.requestWithResponse.Delete(key)
}

// 读取请求 ID 为 key 的命令请求的返回值，
// 同时移除此命令请求
func (c *commandRequestWithResponse) LoadResponseAndDelete(key uuid.UUID) CommandRespond {
	value, exist := c.requestWithResponse.Load(key)
	if !exist {
		return CommandRespond{
			Error:     fmt.Errorf("LoadResponseAndDelete: %v is not recorded", key.String()),
			ErrorType: ErrCommandRequestNotRecord,
		}
	}
	// if key is not exist
	chanGet, normal := value.(chan packet.CommandOutput)
	if !normal {
		return CommandRespond{
			Error:     fmt.Errorf("LoadResponseAndDelete: Failed to convert value into (chan packet.CommandOutput); value = %#v", value),
			ErrorType: ErrCommandRequestConversionFailure,
		}
	}
	// convert data
	select {
	case res := <-chanGet:
		c.requestWithResponse.Delete(key)
		return CommandRespond{Respond: res}
	case <-time.After(CommandRequestDeadLine):
		c.requestWithResponse.Delete(key)
		return CommandRespond{
			Error:     fmt.Errorf(`LoadResponseAndDelete: Request "%v" time out`, key.String()),
			ErrorType: ErrCommandRequestTimeOut,
		}
	}
	// process and return
}
