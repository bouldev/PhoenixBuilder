package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 切换客户端的手持物品栏为 hotBarSlotID；如果 needWaiting 为真，则会等待物品栏切换完成后再返回值
func (g *GlobalAPI) ChangeSelectedHotbarSlot(hotBarSlotID uint8, needWaiting bool) error {
	var got protocol.ItemInstance = protocol.ItemInstance{}
	// init var
	datas, err := g.PacketHandleResult.Inventory.GetItemStackInfo(0, 0)
	// get item contents of window 0
	if err != nil {
		got = protocol.ItemInstance{
			StackNetworkID: 0,
			Stack: protocol.ItemStack{
				ItemType: protocol.ItemType{
					NetworkID:     0,
					MetadataValue: 0,
				},
				BlockRuntimeID: 0,
				Count:          0,
				NBTData:        map[string]interface{}{},
				CanBePlacedOn:  []string(nil),
				CanBreak:       []string(nil),
				HasNetworkID:   false,
			},
		}
	} else {
		got = datas
	}
	// get target item datas
	err = g.WritePacket(&packet.MobEquipment{
		EntityRuntimeID: g.BotRunTimeID,
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
