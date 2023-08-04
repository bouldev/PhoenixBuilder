package GameInterface

import (
	"encoding/gob"
	"fmt"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 将已放入铁砧第一格(注意是第一格)的物品的物品名称修改为 name 并返还到背包中的 slot 处。
// 当且仅当租赁服回应操作结果后此函数再返回值。
//
// 返回值的第一项代表物品名称的修改结果。
// 当发生错误时此参数将始终为 nil
//
// 当遭遇改名失败时，将尝试撤销名称修改请求。
// 如果原有物品栏已被占用，则会尝试将铁砧中的失败品
// 返还到背包中另外一个可用的物品栏。
// 如果背包中所有物品栏都已被占用，则物品会被留在铁砧内。
//
// 部分情况下此函数可能会遇见无法处理的错误，届时程序将抛出严重错误(panic)
func (g *GameInterface) RenameItem(
	name string,
	slot uint8,
) (*AnvilOperationResponse, error) {
	containerOpeningData := g.Resources.Container.GetContainerOpeningData()
	// 取得已打开的容器的数据
	if containerOpeningData == nil {
		return nil, fmt.Errorf("RenameItem: Anvil has not opened")
	}
	// 如果铁砧未被打开
	get, err := g.Resources.Inventory.GetItemStackInfo(
		uint32(containerOpeningData.WindowID),
		1,
	)
	if err != nil {
		return nil, fmt.Errorf("RenameItem: %v", err)
	}
	if get.Stack.NetworkID == 0 {
		return nil, fmt.Errorf("RenameItem: Item provided is air")
	}
	// 得到已放入铁砧的物品的物品数据，
	// 并进行数据检查
	var itemDatas protocol.ItemInstance
	ResourcesControl.DeepCopy(
		&get,
		&itemDatas,
		func() {
			gob.Register(map[string]interface{}{})
			gob.Register([]interface{}{})
		},
	)
	// 将得到的物品数据深拷贝到并 itemDatas 处
	var backup protocol.ItemInstance
	ResourcesControl.DeepCopy(
		&get,
		&backup,
		func() {
			gob.Register(map[string]interface{}{})
			gob.Register([]interface{}{})
		},
	)
	// 备份物品数据到 backup 处
	getOptionalSlot := func() []uint8 {
		filterAns, err := g.Resources.Inventory.ListSlot(0, &[]int32{0})
		if err != nil {
			panic(fmt.Sprintf("RenameItem: %v", err))
		}
		// 筛选出背包中还未被占用实际物品的物品栏
		optionalSlot := []uint8{slot}
		optionalSlot = append(optionalSlot, filterAns...)
		if len(filterAns) <= 0 {
			optionalSlot = []uint8{}
		}
		// optionalSlot 指代被操作物品最终可能出现的位置
		return optionalSlot
		// return
	}
	// 构造一个函数以用于取得背包中的空气物品栏
	revertFunc := func() (*AnvilOperationResponse, error) {
		for _, value := range getOptionalSlot() {
			placeStackRequestAction := protocol.PlaceStackRequestAction{}
			placeStackRequestAction.Source = protocol.StackRequestSlotInfo{
				ContainerID:    0,
				Slot:           1,
				StackNetworkID: backup.StackNetworkID,
			}
			placeStackRequestAction.Destination = protocol.StackRequestSlotInfo{
				ContainerID:    0xc,
				Slot:           value,
				StackNetworkID: 0,
			}
			placeStackRequestAction.Count = byte(backup.Stack.Count)
			// 构造一个新的 placeStackRequestAction 结构体
			resp, err := g.SendItemStackRequestWithResponse(
				&packet.ItemStackRequest{
					Requests: []protocol.ItemStackRequest{
						{
							RequestID: g.Resources.ItemStackOperation.GetNewRequestID(),
							Actions: []protocol.StackRequestAction{
								&placeStackRequestAction,
							},
						},
					},
				},
				[]ItemChangingDetails{
					{
						Details: map[ResourcesControl.ContainerID]ResourcesControl.StackRequestContainerInfo{
							0x0: {
								WindowID: uint32(containerOpeningData.WindowID),
								ChangeResult: map[uint8]protocol.ItemInstance{
									1: AirItem,
								},
							},
							0xc: {
								WindowID: 0,
								ChangeResult: map[uint8]protocol.ItemInstance{
									value: backup,
								},
							},
						},
					},
				},
			)
			// 尝试将被槽位物品还原到背包中的 value 物品栏处
			if err != nil {
				return nil, err
			}
			if resp[0].Status == protocol.ItemStackResponseStatusOK {
				return &AnvilOperationResponse{
					Successful: false,
					Destination: &ItemLocation{
						WindowID:    0,
						ContainerID: 0xc,
						Slot:        value,
					},
				}, nil
			}
			// 如果成功还原的话，那么返回值
		}
		// 尝试把被操作物品从铁砧放回背包中
		return &AnvilOperationResponse{
			Successful: false,
			Destination: &ItemLocation{
				WindowID:    containerOpeningData.WindowID,
				ContainerID: 0,
				Slot:        1,
			},
		}, nil
		// 看起来背包已经满了，我们不得不把物品留在铁砧中
	}
	// 构造一个函数用于处理改名失败时的善后处理
	for _, value := range getOptionalSlot() {
		newRequestID := g.Resources.ItemStackOperation.GetNewRequestID()
		// 请求一个新的 RequestID 用于 ItemStackRequest
		placeStackRequestAction := protocol.PlaceStackRequestAction{}
		placeStackRequestAction.Count = byte(itemDatas.Stack.Count)
		placeStackRequestAction.Source = protocol.StackRequestSlotInfo{
			ContainerID:    0x3c,
			Slot:           0x32,
			StackNetworkID: newRequestID,
		}
		placeStackRequestAction.Destination = protocol.StackRequestSlotInfo{
			ContainerID:    0xc,
			Slot:           value,
			StackNetworkID: 0,
		}
		// 构造一个新的 PlaceStackRequestAction 结构体
		newItemStackRequest := packet.ItemStackRequest{
			Requests: []protocol.ItemStackRequest{
				{
					RequestID: newRequestID,
					Actions: []protocol.StackRequestAction{
						&protocol.CraftRecipeOptionalStackRequestAction{
							RecipeNetworkID:   0,
							FilterStringIndex: 0,
						},
						&protocol.ConsumeStackRequestAction{
							DestroyStackRequestAction: protocol.DestroyStackRequestAction{
								Count: byte(itemDatas.Stack.Count),
								Source: protocol.StackRequestSlotInfo{
									ContainerID:    0,
									Slot:           1,
									StackNetworkID: itemDatas.StackNetworkID,
								},
							},
						},
						&placeStackRequestAction,
					},
					FilterStrings: []string{name},
				},
			},
		}
		// 构造一个新的 ItemStackRequest 结构体
		err = g.Resources.ItemStackOperation.SetItemName(
			&itemDatas,
			name,
		)
		if err != nil {
			return revertFunc()
		}
		// 更新物品数据中的名称字段以用于更新本地库存数据
		_, ok := itemDatas.Stack.NBTData["RepairCost"]
		if !ok {
			itemDatas.Stack.NBTData["RepairCost"] = int32(0)
		}
		// 更新物品数据中的 RepairCost 字段用于更新本地库存数据
		err = g.Resources.ItemStackOperation.WriteRequest(
			newRequestID,
			map[ResourcesControl.ContainerID]ResourcesControl.StackRequestContainerInfo{
				0xc: {
					WindowID: 0,
					ChangeResult: map[uint8]protocol.ItemInstance{
						value: itemDatas,
					},
				},
				0x0: {
					WindowID: uint32(containerOpeningData.WindowID),
					ChangeResult: map[uint8]protocol.ItemInstance{
						1: AirItem,
					},
				},
				0x1: {
					WindowID: uint32(containerOpeningData.WindowID),
					ChangeResult: map[uint8]protocol.ItemInstance{
						2: AirItem,
					},
				},
				0x3c: {
					WindowID: uint32(containerOpeningData.WindowID),
					ChangeResult: map[uint8]protocol.ItemInstance{
						0x32: AirItem,
					},
				},
			},
		)
		if err != nil {
			return revertFunc()
		}
		// 写入请求到等待队列
		err = g.WritePacket(&newItemStackRequest)
		if err != nil {
			panic(fmt.Sprintf("RenameItem: %v", err))
		}
		// 发送物品操作请求
		ans, err := g.Resources.ItemStackOperation.LoadResponseAndDelete(newRequestID)
		if err != nil {
			return revertFunc()
		}
		// 等待租赁服响应物品操作请求并取得物品名称操作结果
		if ans.Status == 0x9 {
			return revertFunc()
		}
		// 此时改名失败，原因是物品的新名称与原始名称重名
		if ans.Status == protocol.ItemStackResponseStatusOK {
			return &AnvilOperationResponse{
				Successful: true,
				Destination: &ItemLocation{
					WindowID:    0,
					ContainerID: 0xc,
					Slot:        value,
				},
			}, nil
		}
		// 当改名成功时
	}
	return revertFunc()
	// 返回值
}
