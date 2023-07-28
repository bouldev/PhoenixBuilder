package GameInterface

import (
	"fmt"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 每个物品操作请求都会使用这样一个结构体，
// 它用于描述单个的物品操作请求中各容器中各槽位的变动情况
type ItemChangingDetails struct {
	// 描述多个库存(容器)中物品的变动结果。
	// 租赁服不会在返回 ItemStackResponce 时返回完整的物品数据，因此需要您提供对应
	// 槽位的更改结果以便于我们依此更新本地存储的库存数据
	Details map[ResourcesControl.ContainerID]ResourcesControl.StackRequestContainerInfo
}

/*
向租赁服发送 ItemStackReuqest 并获取返回值。
request 指代已经构造好的物品操作请求，
details 指代物品操作请求后物品的变动结果，这将用于更新本地库存数据。

注意：
1. 无论 request 中填写的 RequestID 是什么，它们最终会被覆写为正确的值；
2. request.Requests[key] 与 details[key] 一一对应
*/
func (g *GameInterface) SendItemStackRequestWithResponse(
	request *packet.ItemStackRequest,
	details []ItemChangingDetails,
) ([]protocol.ItemStackResponse, error) {
	requestIDList := []int32{}
	res := []protocol.ItemStackResponse{}
	// 初始化
	for range request.Requests {
		requestIDList = append(
			requestIDList,
			g.Resources.ItemStackOperation.GetNewRequestID(),
		)
	}
	for key := range request.Requests {
		requestID := requestIDList[key]
		request.Requests[key].RequestID = requestID
		g.Resources.ItemStackOperation.WriteRequest(
			requestID,
			details[key].Details,
		)
	}
	// 重新设定每个请求的请求 ID 并写入请求到等待队列
	err := g.WritePacket(request)
	if err != nil {
		return nil, fmt.Errorf("SendItemStackRequestWithResponse: %v", err)
	}
	// 发送物品操作请求
	for _, value := range requestIDList {
		got, err := g.Resources.ItemStackOperation.LoadResponseAndDelete(value)
		if err != nil {
			return nil, fmt.Errorf("SendItemStackRequestWithResponse: %v", err)
		}
		res = append(res, got)
	}
	// 等待租赁服回应所有物品操作请求。同时，每当一个请求被响应，就把对应的结果保存下来
	return res, nil
	// 返回值
}
