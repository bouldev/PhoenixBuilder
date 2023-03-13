package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 向租赁服发送 ItemStackReuqest 并获取返回值
func (g *GlobalAPI) SendItemStackRequestWithResponce(request *packet.ItemStackRequest) ([]protocol.ItemStackResponse, error) {
	requestIDList := []int32{}
	ans := []protocol.ItemStackResponse{}
	// 初始化
	for range request.Requests {
		requestIDList = append(requestIDList, g.PacketHandleResult.ItemStackOperation.GetNewRequestID())
	}
	for key := range request.Requests {
		requestID := requestIDList[key]
		request.Requests[key].RequestID = requestID
		g.PacketHandleResult.ItemStackOperation.WriteRequest(requestID)
	}
	// 重新设定每个请求的请求 ID 并写入请求到等待队列
	err := g.WritePacket(request)
	if err != nil {
		return nil, fmt.Errorf("SendItemStackRequestWithResponce: %v", err)
	}
	// 发送物品操作请求
	for _, value := range requestIDList {
		g.PacketHandleResult.ItemStackOperation.AwaitResponce(value)
		got, err := g.PacketHandleResult.ItemStackOperation.LoadResponceAndDelete(value)
		if err != nil {
			return nil, fmt.Errorf("SendItemStackRequestWithResponce: %v", err)
		}
		ans = append(ans, got)
	}
	// 等待租赁服回应所有物品操作请求。同时，每当一个请求被响应，就把对应的结果保存下来
	return ans, nil
	// 返回值
}

/*
将背包中槽位为 inventorySlot 的物品移动到已打开容器的第 containerSlot 槽位，且只移动 moveCount 个物品。
此函数将 containerSlot 处的物品当作空气处理。如果涉及到交换物品等操作，或许您需要使用其他函数。
当且仅当物品操作得到租赁服的响应后，此函数才会返回值
*/
func (g *GlobalAPI) PlaceItemIntoContainer(
	inventorySlot uint8,
	containerSlot uint8,
	moveCount uint8,
) error {
	datas, err := g.PacketHandleResult.Inventory.GetItemStackInfo(0, inventorySlot)
	if err != nil {
		return fmt.Errorf("PlaceItemIntoContainer: %v", err)
	}
	// 取得背包中指定物品栏的物品数据
	placeStackRequestAction := protocol.PlaceStackRequestAction{}
	if moveCount <= uint8(datas.Stack.Count) {
		placeStackRequestAction.Count = moveCount
	} else {
		placeStackRequestAction.Count = uint8(datas.Stack.Count)
	}
	// 得到欲移动的物品数量
	placeStackRequestAction.Source = protocol.StackRequestSlotInfo{
		ContainerID:    12,
		Slot:           inventorySlot,
		StackNetworkID: datas.StackNetworkID,
	}
	placeStackRequestAction.Destination = protocol.StackRequestSlotInfo{
		ContainerID:    g.PacketHandleResult.ContainerResources.GetContainerOpenDatas().ContainerType,
		Slot:           containerSlot,
		StackNetworkID: 0,
	}
	// 前置准备
	ans, err := g.SendItemStackRequestWithResponce(&packet.ItemStackRequest{
		Requests: []protocol.ItemStackRequest{
			{
				Actions: []protocol.StackRequestAction{
					&placeStackRequestAction,
				},
				FilterStrings: []string{},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("PlaceItemIntoContainer: %v", err)
	}
	// 发送物品操作请求
	if ans[0].Status != 0 {
		return fmt.Errorf("PlaceItemIntoContainer: Operation %v is canceled, and the errorCode(status) is %v; inventorySlot = %v, containerSlot = %v, moveCount = %v", ans[0].RequestID, ans[0].Status, inventorySlot, containerSlot, moveCount)
	}
	// 当操作失败时
	return nil
	// 返回值
}
