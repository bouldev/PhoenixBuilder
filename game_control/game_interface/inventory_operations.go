package GameInterface

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

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
如需要关闭已打开的背包，请直接使用函数 CloseContainer 。

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
