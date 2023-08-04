package GameInterface

import (
	"fmt"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"
)

// 将快捷栏 hotBarSlot 处的可堆叠物品复制为 requestCount 个。
// 此操作需要占用一个方块，因此 blockPos 用于描述此方块的坐标。
func (g *GameInterface) CopyItem(
	hotBarSlot uint8,
	blockPos [3]int32,
	requestCount uint8,
) error {
	openConatiner := func() error {
		success, err := g.OpenContainer(
			blockPos,
			"minecraft:barrel",
			map[string]interface{}{
				"facing_direction": int32(0),
				"open_bit":         byte(0),
			},
			5,
		)
		if err != nil {
			return fmt.Errorf("openConatiner: %v", err)
		}
		if !success {
			return fmt.Errorf("openConatiner: Failed to open the barrel")
		}
		// 打开已生成的木桶
		listener, packets := g.Resources.Listener.CreateNewListen([]uint32{packet.IDInventoryContent}, 1)
		select {
		case <-packets:
		case <-time.After(ItemCopyWaitContainerDataComingDeadLine):
		}
		err = g.Resources.Listener.StopAndDestroy(listener)
		if err != nil {
			return fmt.Errorf("openConatiner: %v", err)
		}
		return nil
		// 等待数据同步
	}
	// 打开已放置的木桶并等待数据同步完成
	subFunc := func(
		source ItemLocation,
		destination ItemLocation,
	) (bool, error) {
		var moveCounts int8
		var result uint16
		// 初始化
		sourceData, err := g.Resources.Inventory.GetItemStackInfo(uint32(source.WindowID), source.Slot)
		if err != nil {
			return false, fmt.Errorf("subFunc: %v", err)
		}
		destinationData, err := g.Resources.Inventory.GetItemStackInfo(uint32(destination.WindowID), destination.Slot)
		if err != nil {
			return false, fmt.Errorf("subFunc: %v", err)
		}
		// 获取物品数据
		sum := sourceData.Stack.Count + destinationData.Stack.Count
		if sum > uint16(requestCount) {
			moveCounts = int8(requestCount - uint8(destinationData.Stack.Count))
			result = uint16(requestCount)
		} else {
			moveCounts = int8(sourceData.Stack.Count)
			result = sum
		}
		// 确定应当移动的物品数量和移动成功后的结果
		if sourceData.Stack.NetworkID == 0 {
			return false, fmt.Errorf("subFunc: Unknown item loss occurred")
		}
		if moveCounts <= 0 {
			return true, nil
		}
		// 数据检查
		resp, err := g.MoveItem(
			source,
			destination,
			uint8(moveCounts),
			AirItem,
			protocol.ItemInstance{
				Stack: protocol.ItemStack{
					ItemType:       sourceData.Stack.ItemType,
					BlockRuntimeID: sourceData.Stack.BlockRuntimeID,
					Count:          result,
					NBTData:        sourceData.Stack.NBTData,
					CanBePlacedOn:  sourceData.Stack.CanBePlacedOn,
					CanBreak:       sourceData.Stack.CanBreak,
					HasNetworkID:   sourceData.Stack.HasNetworkID,
				},
			},
		)
		if err != nil {
			return false, fmt.Errorf("subFunc: %v", err)
		}
		if resp[0].Status != protocol.ItemStackResponseStatusOK {
			return false, fmt.Errorf("subFunc: The request was rejected by the remote server; resp[0] = %#v", resp[0])
		}
		// 合并 source 处的物品到 destination 处
		if sum >= uint16(requestCount) {
			return true, nil
		}
		return false, nil
		// 返回值
	}
	// 将 source 处的物品合并到 destination 处
	// 返回的布尔值代表合并后的物品总数是否已达到要求的数目
	singleCodeBlock := func() (bool, error) {
		err := openConatiner()
		defer g.CloseContainer()
		if err != nil {
			return false, fmt.Errorf("singleCodeBlock: %v", err)
		}
		// 打开木桶并等待数据同步
		itemOnHotBarSlot := ItemLocation{
			WindowID:    0,
			ContainerID: 0xc,
			Slot:        hotBarSlot,
		}
		itemOnConatiner := ItemLocation{
			WindowID:    g.Resources.Container.GetContainerOpeningData().WindowID,
			ContainerID: 58,
			Slot:        13,
		}
		// 初始化
		_, err = subFunc(itemOnHotBarSlot, itemOnConatiner)
		if err != nil {
			return false, fmt.Errorf("singleCodeBlock: %v", err)
		}
		// 快捷栏处的物品 --合并到--> 木桶的第 14 个槽位
		uniqueId, err := g.BackupStructure(
			MCStructure{
				BeginX: blockPos[0],
				BeginY: blockPos[1],
				BeginZ: blockPos[2],
				SizeX:  1,
				SizeY:  1,
				SizeZ:  1,
			},
		)
		if err != nil {
			return false, fmt.Errorf("singleCodeBlock: %v", err)
		}
		defer g.RevertStructure(uniqueId, blockPos)
		// 备份木桶
		err = g.ReplaceItemInInventory(
			TargetMySelf,
			ItemGenerateLocation{
				Path: "slot.hotbar",
				Slot: hotBarSlot,
			},
			types.ChestSlot{
				Name:   "air",
				Count:  1,
				Damage: 0,
			},
			"",
			true,
		)
		if err != nil {
			return false, fmt.Errorf("singleCodeBlock: %v", err)
		}
		needBreak, err := subFunc(itemOnConatiner, itemOnHotBarSlot)
		if err != nil {
			return false, fmt.Errorf("singleCodeBlock: %v", err)
		}
		if needBreak {
			return true, nil
		}
		// 将已放入木桶的物品取回
		return false, nil
		// 返回值
	}
	// 将 hotBarSlot 处的物品变为双倍。
	// 返回的布尔值代表本次翻倍后的物品总数是否已达到要求的数目，
	// 这被用于下方 for 循环代码块的终止条件。
	{
		if requestCount == 0 {
			return nil
		}
		// 此时视为无需复制物品
		holder := g.Resources.Container.Occupy()
		defer g.Resources.Container.Release(holder)
		// 占用容器资源
		err := g.SendSettingsCommand(
			fmt.Sprintf("tp %d %d %d", blockPos[0], blockPos[1], blockPos[2]),
			true,
		)
		if err != nil {
			return fmt.Errorf("CopyItem: %v", err)
		}
		// 传送机器人到 blockPos
		err = g.SetBlockAsync(blockPos, "air", "[]")
		if err != nil {
			return fmt.Errorf("CopyItem: %v", err)
		}
		err = g.SetBlock(blockPos, "barrel", `["facing_direction": 0]`)
		if err != nil {
			return fmt.Errorf("CopyItem: %v", err)
		}
		err = openConatiner()
		if err != nil {
			return fmt.Errorf("CopyItem: %v", err)
		}
		defer g.CloseContainer()
		// 生成木桶，然后打开其并等待数据同步
		for {
			stop, err := singleCodeBlock()
			if err != nil {
				return fmt.Errorf("CopyItem: %v", err)
			}
			if stop {
				break
			}
		}
		// 复制物品
		return nil
		// 返回值
	}
	// 复制物品为指定数目
}
