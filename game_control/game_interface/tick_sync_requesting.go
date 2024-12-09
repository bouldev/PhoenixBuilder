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
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 用于获取当前的游戏刻。
// 此操作不会被立即完成，
// 因为它需要请求一个数据包
func (g *GameInterface) GetCurrentTick() (int64, error) {
	uniqueId := ResourcesControl.GenerateUUID()
	// get a new uuid
	err := g.Resources.Others.WriteGameTickRequest(uniqueId)
	if err != nil {
		return 0, fmt.Errorf("GetCurrentTick: %v", err)
	}
	// write request
	err = g.WritePacket(&packet.TickSync{
		ClientRequestTimestamp:   0,
		ServerReceptionTimestamp: 0,
	})
	if err != nil {
		return 0, fmt.Errorf("GetCurrentTick: %v", err)
	}
	// send packet
	ans, err := g.Resources.Others.LoadTickSyncResponse(uniqueId)
	if err != nil {
		return 0, fmt.Errorf("GetCurrentTick: %v", err)
	}
	return ans, nil
	// load responce and return
}
