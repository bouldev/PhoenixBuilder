package GameInterface

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 切换客户端的手持物品栏为 hotBarSlotID 。
// 若提供的 hotBarSlotID 大于 8 ，则会重定向为 0
func (g *GameInterface) ChangeSelectedHotbarSlot(hotbarSlotID uint8) error {
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

/*
打开背包。
返回值的第一项代表执行结果，为真时背包被成功打开，否则反之。
如需要关闭已打开的背包，请直接使用函数 CloseContainer .

请确保打开前占用了容器资源，否则会造成程序 panic 。
*/
func (g *GameInterface) OpenInventory() (bool, error) {
	g.Resources.Container.AwaitChangesBeforeSendingPacket()
	// await responce before send packet
	err := g.WritePacket(&packet.Interact{
		ActionType:            packet.InteractActionOpenInventory,
		TargetEntityRuntimeID: g.ClientInfo.EntityRuntimeID,
	})
	if err != nil {
		return false, fmt.Errorf("OpenInventory: %v", err)
	}
	// open inventory
	g.Resources.Container.AwaitChangesAfterSendingPacket()
	// wait changes
	if g.Resources.Container.GetContainerOpeningData() == nil {
		return false, nil
	}
	// if unsuccess
	return true, nil
	// return
}
