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

// 请求 request 代表的结构请求并获取与之对应的响应体。
// 当且仅当租赁服响应结构请求时本函数才会返回值。
//
// 请确保在使用此函数前占用了结构资源，否则这将导致程序 panic
func (g *GameInterface) SendStructureRequestWithResponse(
	request *packet.StructureTemplateDataRequest,
) (packet.StructureTemplateDataResponse, error) {
	g.Resources.Structure.WriteRequest()
	// prepare
	err := g.WritePacket(request)
	if err != nil {
		return packet.StructureTemplateDataResponse{}, fmt.Errorf("SendStructureRequestWithResponse: %v", err)
	}
	// send packet
	return g.Resources.Structure.LoadResponse(), nil
	// load response and return
}
