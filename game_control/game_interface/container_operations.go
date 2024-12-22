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

/*
打开 pos 处名为 blockName 且方块状态为 blockStates 的容器。
hotBarSlotID 字段代表玩家此时手持的物品栏，
因为打开容器实际上是一次方块点击事件。
返回值的第一项代表执行结果，为真时容器被成功打开，否则反之。

容器不一定总能打开，可能该容器已被移除或机器人已被移动。
因此，单次打开操作在抵达最长截止时间后将会在内部被验证为超时，
此时将会重新提交一次容器打开操作，
直到总操作次数抵达 ContainerOperationsReTryMaximumCounts 时止。

请确保在使用此函数前占用了容器资源，否则会造成程序 panic
*/
func (g *GameInterface) OpenContainer(
	pos [3]int32,
	blockName string,
	blockStates map[string]interface{},
	hotBarSlotID uint8,
) (bool, error) {
	if g.Resources.Container.GetContainerOpeningData() != nil {
		return false, ErrContainerHasBeenOpened
	}
	// if the container has been opened
	for i := 0; i < ContainerOperationsReTryMaximumCounts; i++ {
		g.Resources.Container.AwaitChangesBeforeSendingPacket()
		// await responce before send packet
		err := g.ClickBlock(
			UseItemOnBlocks{
				HotbarSlotID: hotBarSlotID,
				BlockPos:     pos,
				BlockName:    blockName,
				BlockStates:  blockStates,
			},
		)
		if err != nil {
			return false, fmt.Errorf("OpenContainer: %v", err)
		}
		// open container
		g.Resources.Container.AwaitChangesAfterSendingPacket()
		// wait changes
		if g.Resources.Container.GetContainerOpeningData() == nil {
			continue
		}
		// if unsuccess
		return true, nil
		// return
	}
	// open container.
	// try a maximum of ContainerOperationsReTryMaximumCounts times
	return false, nil
	// return
}

/*
关闭已经打开的容器，且只有当容器被关闭后才会返回值。
您应该确保容器被关闭后，对应的容器公用资源被释放。

返回值的第一项代表执行结果，为真时容器被成功关闭，否则反之。

容器不一定总能关闭，可能租赁服已经卡死。
因此，单次关闭操作在抵达最长截止时间后将会在内部被验证为超时，
此时将会重新提交一次容器打开操作，
直到总操作次数抵达 ContainerOperationsReTryMaximumCounts 时止。
*/
func (g *GameInterface) CloseContainer() (bool, error) {
	if g.Resources.Container.GetContainerOpeningData() == nil {
		return false, ErrContainerNerverOpened
	}
	// if the container is not opened
	for i := 0; i < ContainerOperationsReTryMaximumCounts; i++ {
		g.Resources.Container.AwaitChangesBeforeSendingPacket()
		// await responce before send packet
		err := g.WritePacket(&packet.ContainerClose{
			WindowID:   g.Resources.Container.GetContainerOpeningData().WindowID,
			ServerSide: false,
		})
		if err != nil {
			return false, fmt.Errorf("CloseContainer: %v", err)
		}
		// close container
		g.Resources.Container.AwaitChangesAfterSendingPacket()
		// wait changes
		if g.Resources.Container.GetContainerClosingData() == nil {
			continue
		}
		// if unsuccess
		return true, nil
		// return
	}
	// close container.
	// try a maximum of ContainerOperationsReTryMaximumCounts times
	return false, nil
	// return
}
