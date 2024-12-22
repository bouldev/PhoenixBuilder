package ResourcesControl

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

import "phoenixbuilder/minecraft/protocol/packet"

// 提交结构请求
func (m *mcstructure) WriteRequest() {
	m.resp = make(chan packet.StructureTemplateDataResponse, 1)
}

// 向结构请求写入返回值 resp 。
// 属于私有实现。
func (m *mcstructure) writeResponse(data packet.StructureTemplateDataResponse) {
	m.resp <- data
	close(m.resp)
}

// 从管道读取结构请求的返回值
func (m *mcstructure) LoadResponse() packet.StructureTemplateDataResponse {
	return <-m.resp
}
