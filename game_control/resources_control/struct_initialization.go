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

import (
	"phoenixbuilder/fastbuilder/generics"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/google/uuid"
)

/*
初始化 Resources 结构体并返回一个函数用于更新资源。

!!! 重要 !!!
请务必保证在单个服务器连接下，此函数至多被调用一次，
否则会发生无法解决的冲突性问题
*/
func (r *Resources) Init() func(pk *packet.Packet) {
	*r = Resources{
		Command: command_request_with_response{
			ai_command_resp: nil,
			request_lock:    sync.RWMutex{},
			request:         orderedmap.NewOrderedMap[uuid.UUID, CommandRequestOptions](),
			response:        generics.SyncMap[uuid.UUID, *CommandRespond]{},
			signal:          generics.SyncMap[uuid.UUID, chan uint8]{},
		},
		Inventory: inventory_contents{
			lock_down: sync.RWMutex{},
			data:      generics.SyncMap[uint32, *generics.SyncMap[uint8, protocol.ItemInstance]]{},
		},
		ItemStackOperation: item_stack_request_with_response{
			request_with_response: generics.SyncMap[int32, singleitem_stack_request_with_response]{},
			current_request_id:    1,
		},
		Container: container{
			lock_down:              sync.RWMutex{},
			container_opening_data: nil,
			container_closing_data: nil,
			responded:              make(chan struct{}, 1),
			resources_occupy: resources_occupy{
				lock_down: sync.Mutex{},
				holder:    "",
			},
		},
		Structure: mcstructure{
			resources_occupy: resources_occupy{
				lock_down: sync.Mutex{},
				holder:    "",
			},
			resp: make(chan packet.StructureTemplateDataResponse, 1),
		},
		Listener: packet_listener{
			listener_with_data: generics.SyncMap[uuid.UUID, single_listen]{},
		},
		Others: others{
			current_tick_request_with_resp: generics.SyncMap[uuid.UUID, chan int64]{},
		},
	}
	// init struct
	return r.handlePacket
	// return
}
