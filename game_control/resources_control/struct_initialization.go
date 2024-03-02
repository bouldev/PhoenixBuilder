package ResourcesControl

import (
	"phoenixbuilder/fastbuilder/sync_map"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"

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
		Command: commandRequestWithResponse{
			request:  sync_map.Map[uuid.UUID, CommandRequestOptions]{},
			response: sync_map.Map[uuid.UUID, chan packet.CommandOutput]{},
		},
		Inventory: inventoryContents{
			lockDown: sync.RWMutex{},
			datas:    sync_map.Map[uint32, *sync_map.Map[uint8, protocol.ItemInstance]]{},
		},
		ItemStackOperation: itemStackRequestWithResponse{
			requestWithResponse: sync_map.Map[int32, singleItemStackRequestWithResponse]{},
			currentRequestID:    1,
		},
		Container: container{
			lockDown:             sync.RWMutex{},
			containerOpeningData: nil,
			containerClosingData: nil,
			responded:            make(chan struct{}, 1),
			resourcesOccupy: resourcesOccupy{
				lockDown: sync.Mutex{},
				holder:   "",
			},
		},
		Structure: mcstructure{
			resourcesOccupy: resourcesOccupy{
				lockDown: sync.Mutex{},
				holder:   "",
			},
			resp: make(chan packet.StructureTemplateDataResponse, 1),
		},
		Listener: packetListener{
			listenerWithData: sync_map.Map[uuid.UUID, singleListen]{},
		},
		Others: others{
			currentTickRequestWithResp: sync_map.Map[uuid.UUID, chan int64]{},
		},
	}
	// init struct
	return r.handlePacket
	// return
}
