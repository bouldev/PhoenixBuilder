package GameInterface

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 在使用手持物品对方块进行操作时的通用结构体
type UseItemOnBlocks struct {
	HotbarSlotID uint8                  // 指代机器人当前已选择的快捷栏编号
	BlockPos     [3]int32               // 指代被操作方块的位置
	BlockName    string                 // 指代被操作方块的名称(需要包含命名空间)
	BlockStates  map[string]interface{} // 指代被操作方块的方块状态
}

/*
让客户端点击 request 所指代的方块。

你可以对容器使用这样的操作，这会使得容器被打开。

你亦可以对物品展示框使用这样的操作，
这会使得物品被放入或令展示框内的物品旋转。

此函数不会自动切换物品栏，也不会等待租赁服响应更改
*/
func (g *GameInterface) ClickBlock(request UseItemOnBlocks) error {
	blockRuntimeID, err := blockStatesToNEMCRuntimeID(
		request.BlockName,
		request.BlockStates,
	)
	if err != nil {
		return fmt.Errorf("ClickBlock: %v", err)
	}
	// get block runtime id
	datas, err := g.Resources.Inventory.GetItemStackInfo(0, request.HotbarSlotID)
	if err != nil {
		return fmt.Errorf("ClickBlock: %v", err)
	}
	// get datas of the target item stack
	err = g.WritePacket(&packet.InventoryTransaction{
		LegacyRequestID:    0,
		LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
		Actions:            []protocol.InventoryAction{},
		TransactionData: &protocol.UseItemTransactionData{
			LegacyRequestID:    0,
			LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
			Actions:            []protocol.InventoryAction(nil),
			ActionType:         protocol.UseItemActionClickBlock,
			BlockPosition:      request.BlockPos,
			HotBarSlot:         int32(request.HotbarSlotID),
			HeldItem:           datas,
			BlockRuntimeID:     blockRuntimeID,
		},
	})
	if err != nil {
		return fmt.Errorf("ClickBlock: %v", err)
	}
	err = g.WritePacket(&packet.PlayerAction{
		EntityRuntimeID: g.ClientInfo.EntityRuntimeID,
		ActionType:      protocol.PlayerActionStartBuildingBlock,
		BlockPosition:   request.BlockPos,
	})
	if err != nil {
		return fmt.Errorf("ClickBlock: %v", err)
	}
	// send packet
	return nil
	// return
}

/*
让客户端创建一个新方块。

request 指代实际被点击的方块，但这并不代表新方块被创建的位置。
我们通过点击 request 处的方块，并指定点击的面为 blockFace ，
然后租赁服根据这些信息，在另外相应的位置创建这些新的方块。

此函数不会自动切换物品栏，也不会等待租赁服响应更改
*/
func (g *GameInterface) PlaceBlock(
	request UseItemOnBlocks,
	blockFace int32,
) error {
	blockRuntimeID, err := blockStatesToNEMCRuntimeID(
		request.BlockName,
		request.BlockStates,
	)
	if err != nil {
		return fmt.Errorf("PlaceBlock: %v", err)
	}
	// get block runtime id
	datas, err := g.Resources.Inventory.GetItemStackInfo(0, request.HotbarSlotID)
	if err != nil {
		return fmt.Errorf("PlaceBlock: %v", err)
	}
	// get datas of the target item stack
	err = g.WritePacket(&packet.PlayerAction{
		EntityRuntimeID: g.ClientInfo.EntityRuntimeID,
		ActionType:      protocol.PlayerActionStartItemUseOn,
		BlockPosition:   request.BlockPos,
	})
	if err != nil {
		return fmt.Errorf("PlaceBlock: %v", err)
	}
	err = g.WritePacket(&packet.InventoryTransaction{
		LegacyRequestID:    0,
		LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
		Actions:            []protocol.InventoryAction{},
		TransactionData: &protocol.UseItemTransactionData{
			LegacyRequestID:    0,
			LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
			Actions:            []protocol.InventoryAction(nil),
			ActionType:         protocol.UseItemActionClickBlock,
			BlockPosition:      request.BlockPos,
			BlockFace:          blockFace,
			HotBarSlot:         int32(request.HotbarSlotID),
			HeldItem:           datas,
			BlockRuntimeID:     blockRuntimeID,
		},
	})
	if err != nil {
		return fmt.Errorf("PlaceBlock: %v", err)
	}
	err = g.WritePacket(&packet.PlayerAction{
		EntityRuntimeID: g.ClientInfo.EntityRuntimeID,
		ActionType:      protocol.PlayerActionStopItemUseOn,
		BlockPosition:   request.BlockPos,
	})
	if err != nil {
		return fmt.Errorf("PlaceBlock: %v", err)
	}
	// send packet
	return nil
	// return
}

// 使用快捷栏 hotbarSlotID 进行一次空点击操作。
// 此函数不会自动切换物品栏，也不会等待租赁服响应更改
func (g *GameInterface) ClickAir(hotbarSlotID uint8) error {
	itemData, err := g.Resources.Inventory.GetItemStackInfo(0, hotbarSlotID)
	if err != nil {
		return fmt.Errorf("ClickAir: %v", err)
	}
	// get datas of the target item stack
	err = g.WritePacket(
		&packet.InventoryTransaction{
			TransactionData: &protocol.UseItemTransactionData{
				ActionType: protocol.UseItemActionClickAir,
				HotBarSlot: int32(hotbarSlotID),
				HeldItem:   itemData,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("ClickAir: %v", err)
	}
	// send packet
	return nil
	// return
}
