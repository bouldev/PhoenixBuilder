package GlobalAPI

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 切换客户端的手持物品栏为 hotBarSlotID 。
// 若提供的 hotBarSlotID 大于 8 ，则会重定向为 0
func (g *GlobalAPI) ChangeSelectedHotbarSlot(hotbarSlotID uint8) error {
	if hotbarSlotID > 8 {
		hotbarSlotID = 0
	}
	// init var
	err := g.WritePacket(&packet.PlayerHotBar{
		SelectedHotBarSlot: uint32(hotbarSlotID),
		WindowID:           0,
		SelectHotBarSlot:   true,
	})
	if err != nil {
		return fmt.Errorf("ChangeSelectedHotbarSlot: %v", err)
	}
	// change selected hotbar slot
	return nil
	// return
}
