package GlobalAPI

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 切换客户端的手持物品栏为 hotBarSlotID 。
// 若提供的 hotBarSlotID 大于 8 ，则会重定向为 0 。
// 如果 needWaiting 为真，则会等待物品栏切换完成后再返回值
func (g *GlobalAPI) ChangeSelectedHotbarSlot(hotBarSlotID uint8, needWaiting bool) error {
	var got protocol.ItemInstance = protocol.ItemInstance{}
	if hotBarSlotID > 8 {
		hotBarSlotID = 0
	}
	// init var
	datas, err := g.Resources.Inventory.GetItemStackInfo(0, hotBarSlotID)
	// get item contents of window 0(inventory)
	if err != nil {
		got = AirItem
	} else {
		got = datas
	}
	// get target item datas
	err = g.WritePacket(&packet.MobEquipment{
		EntityRuntimeID: g.BotInfo.BotRunTimeID,
		NewItem:         got,
		InventorySlot:   hotBarSlotID,
		HotBarSlot:      hotBarSlotID,
		WindowID:        protocol.WindowIDInventory,
	})
	if err != nil {
		return fmt.Errorf("ChangeSelectedHotbarSlot: %v", err)
	}
	// change selected hotbar slot
	if needWaiting {
		_, err = g.SendWSCommandWithResponce("list")
		if err != nil {
			return fmt.Errorf("ChangeSelectedHotbarSlot: %v", err)
		}
	}
	// wait slot changes
	return nil
	// return
}
