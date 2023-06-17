package GlobalAPI

import (
	"fmt"
	"phoenixbuilder/GameControl/ResourcesControlCenter"
	"phoenixbuilder/fastbuilder/mcstructure"
	"phoenixbuilder/minecraft/protocol"
)

// 描述铁砧操作的操作结果
type AnvilOperationResponce struct {
	// 指代操作结果，为真时代表成功，否则反之
	SuccessStates bool
	// 指代被操作物品的最终位置，可能不存在。
	// 如果不存在，则代表物品已被丢出
	Destination *ItemLocation
}

// 使用铁砧修改物品名称时会被使用的结构体
type AnvilChangeItemName struct {
	Slot uint8  // 被修改物品在背包所在的槽位
	Name string // 要修改的目标名称
}

/*
在 pos 处放置一个方块状态为 blockStates 的铁砧，
并使用快捷栏 hotBarSlotID 打开铁砧，
然后依次执行 request 列表中的物品名称修改请求。

若提供的 hotBarSlotID 大于 8 ，则会重定向为 0 。

返回值 []AnvilOperationResponce 代表 request 中每个请求的操作结果，
它们一一对应，且为真时代表成功改名。

如果改名时游戏模式不是创造，或者经验值不足，或者提供的新物品名称与原始值相同，
或者尝试修改一个无法移动到铁砧的物品，那么都会遭到租赁服的拒绝。

另，如果背包已满导致无法把物品放回背包，则我们将尝试把它直接从铁砧丢出。
当然，此函数在执行时会自动更换客户端的游戏模式为创造，因此您无需再手动操作一次游戏模式
*/
func (g *GlobalAPI) ChangeItemNameByUsingAnvil(
	pos [3]int32,
	blockStates string,
	hotBarSlotID uint8,
	request []AnvilChangeItemName,
) ([]AnvilOperationResponce, error) {
	ans := []AnvilOperationResponce{}
	// 初始化
	err := g.SendSettingsCommand("gamemode 1", true)
	if err != nil {
		return []AnvilOperationResponce{}, fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 更换游戏模式为创造
	uniqueId, correctPos, err := g.GenerateNewAnvil(pos, blockStates)
	if err != nil {
		return []AnvilOperationResponce{}, fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 尝试生成一个铁砧并附带承重方块
	_, err = g.SendWSCommandWithResponce(fmt.Sprintf("tp %d %d %d", correctPos[0], correctPos[1], correctPos[2]))
	if err != nil {
		return []AnvilOperationResponce{}, fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 传送机器人到铁砧处
	holder := g.Resources.Container.Occupy()
	defer g.Resources.Container.Release(holder)
	// 获取容器资源
	got, err := mcstructure.ParseStringNBT(blockStates, true)
	if err != nil {
		return []AnvilOperationResponce{}, fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	blockStatesMap, normal := got.(map[string]interface{})
	if !normal {
		return []AnvilOperationResponce{}, fmt.Errorf("ChangeItemNameByUsingAnvil: Could not convert got into map[string]interface{}; got = %#v", got)
	}
	// 获取要求放置的铁砧的方块状态
	err = g.ChangeSelectedHotbarSlot(hotBarSlotID, true)
	if err != nil {
		return []AnvilOperationResponce{}, fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	// 切换手持物品栏
	sucessStates, err := g.OpenContainer(correctPos, "minecraft:anvil", blockStatesMap, hotBarSlotID)
	if err != nil {
		return []AnvilOperationResponce{}, fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
	}
	if !sucessStates {
		return []AnvilOperationResponce{}, fmt.Errorf("ChangeItemNameByUsingAnvil: Failed to open the anvil block on %v", correctPos)
	}
	// 打开铁砧
	defer func() {
		g.CloseContainer()
		// 关闭铁砧
		g.RevertBlocks(uniqueId, correctPos)
		// 恢复铁砧下方的承重方块为原本方块
	}()
	// 退出时应该被调用的函数
	for _, value := range request {
		datas, err := g.Resources.Inventory.GetItemStackInfo(0, value.Slot)
		if err != nil || datas.Stack.ItemType.NetworkID == 0 {
			ans = append(ans, AnvilOperationResponce{
				SuccessStates: false,
				Destination: &ItemLocation{
					WindowID:    0,
					ContainerID: 0xc,
					Slot:        value.Slot,
				},
			})
			continue
		}
		// 获取被改物品的相关信息。
		// 如果发生了错误或指定的物品为空气，
		// 则会跳过这个物品
		containerOpenDatas := g.Resources.Container.GetContainerOpenDatas()
		// 获取已打开的容器的数据
		if containerOpenDatas == nil {
			return ans, fmt.Errorf("ChangeItemNameByUsingAnvil: Anvil have been closed")
		}
		// 确保容器未被关闭
		resp, err := g.MoveItem(
			ItemLocation{
				WindowID:    0,
				ContainerID: 0xc,
				Slot:        value.Slot,
			},
			ItemLocation{
				WindowID:    int16(containerOpenDatas.WindowID),
				ContainerID: 0x0,
				Slot:        1,
			},
			ItemChangeDetails{
				details: map[ResourcesControlCenter.ContainerID]ResourcesControlCenter.StackRequestContainerInfo{
					0xc: {
						WindowID: 0,
						ChangeResult: map[uint8]protocol.ItemInstance{
							value.Slot: AirItem,
						},
					},
					0x0: {
						WindowID: uint32(containerOpenDatas.WindowID),
						ChangeResult: map[uint8]protocol.ItemInstance{
							1: datas,
						},
					},
				},
			},
			uint8(datas.Stack.Count),
		)
		if err != nil {
			return ans, fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
		}
		if resp[0].Status != protocol.ItemStackResponseStatusOK {
			ans = append(ans, AnvilOperationResponce{
				SuccessStates: false,
				Destination: &ItemLocation{
					WindowID:    0,
					ContainerID: 0xc,
					Slot:        value.Slot,
				},
			})
			continue
		}
		// 移动物品到铁砧
		backup, err := g.Resources.Inventory.GetItemStackInfo(uint32(containerOpenDatas.WindowID), 1)
		if err != nil {
			panic(fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err))
		}
		// 备份物品数据
		anvilOperationResp, err := g.ChangeItemName(value.Name, value.Slot)
		if err != nil {
			return ans, fmt.Errorf("ChangeItemNameByUsingAnvil: %v", err)
		}
		// 修改物品名称
		if anvilOperationResp.Destination.ContainerID == 0 {
			successStates, err := g.DropItemAll(
				protocol.StackRequestSlotInfo{
					ContainerID:    0,
					Slot:           1,
					StackNetworkID: backup.StackNetworkID,
				},
				uint32(containerOpenDatas.WindowID),
			)
			if err != nil {
				panic(fmt.Sprintf("ChangeItemNameByUsingAnvil: %v", err))
			}
			if !successStates {
				panic("ChangeItemNameByUsingAnvil: Failure to recover, and we have no choice but to panic this program")
			}
			anvilOperationResp.Destination = nil
		}
		// 如果物品被滞留在了铁砧，那么尝试将物品丢出。
		// 只有当背包已满时才会发生滞留现象
		ans = append(ans, *anvilOperationResp)
		// 提交子结果
	}
	// 修改物品名称
	return ans, nil
	// 返回值
}
