package ResourcesControlCenter

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// 提交请求 ID 为 key 的命令请求
func (c *commandRequestWithResponce) WriteRequest(key uuid.UUID) error {
	_, exist := c.requestWithResponce.Load(key)
	if exist {
		return fmt.Errorf("WriteRequest: %v has already existed", key.String())
	}
	// if key has already exist
	c.requestWithResponce.Store(key, make(chan packet.CommandOutput, 1))
	return nil
	// return
}

// 尝试向请求 ID 为 key 的命令请求写入返回值 resp 。
// 属于私有实现。
// 如果 key 不存在，亦不会返回错误。
func (c *commandRequestWithResponce) tryToWriteResponce(
	key uuid.UUID,
	resp packet.CommandOutput,
) error {
	value, exist := c.requestWithResponce.Load(key)
	if !exist {
		return nil
	}
	// if key is not exist
	chanGet, normal := value.(chan packet.CommandOutput)
	if !normal {
		return fmt.Errorf("tryToWriteResponce: Failed to convert value into (chan packet.CommandOutput); value = %#v", value)
	}
	// convert data
	chanGet <- resp
	close(chanGet)
	return nil
	// return
}

// 读取请求 ID 为 key 的命令请求的返回值，
// 同时移除此命令请求
func (c *commandRequestWithResponce) LoadResponceAndDelete(key uuid.UUID) (packet.CommandOutput, error) {
	value, exist := c.requestWithResponce.Load(key)
	if !exist {
		return packet.CommandOutput{}, fmt.Errorf("LoadResponceAndDelete: %v is not recorded", key.String())
	}
	// if key is not exist
	chanGet, normal := value.(chan packet.CommandOutput)
	if !normal {
		return packet.CommandOutput{}, fmt.Errorf("LoadResponceAndDelete: Failed to convert value into (chan packet.CommandOutput); value = %#v", value)
	}
	// convert data
	ans := <-chanGet
	c.requestWithResponce.Delete(key)
	return ans, nil
	// return
}
