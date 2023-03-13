package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// 向租赁服发送 WS 命令且获取返回值
func (g *GlobalAPI) SendWSCommandWithResponce(command string) (packet.CommandOutput, error) {
	uniqueId, err := uuid.NewUUID()
	if err != nil || uniqueId == uuid.Nil {
		resp, err := g.SendWSCommandWithResponce(command)
		if err != nil {
			return packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponce: %v", err)
		}
		return resp, nil
	}
	err = g.PacketHandleResult.commandDatas.writeRequest(uniqueId)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponce: %v", err)
	}
	// 写入请求到等待队列
	err = g.SendWSCommand(command, uniqueId)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponce: %v", err)
	}
	// 发送命令
	g.PacketHandleResult.commandDatas.awaitResponce(uniqueId)
	// 等待租赁服响应命令请求
	ans, err := g.PacketHandleResult.commandDatas.loadResponceAndDelete(uniqueId)
	if err != nil {
		return packet.CommandOutput{}, fmt.Errorf("SendWSCommandWithResponce: %v", err)
	}
	// 取得命令请求的返回值
	return ans, nil
	// 返回值
}
