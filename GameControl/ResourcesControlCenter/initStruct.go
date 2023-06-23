package ResourcesControlCenter

import (
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
)

/*
初始化 Resources 结构体并返回一个函数用于更新资源。

!!! 重要 !!!
请务必保证在单个服务器连接下，此函数至多被调用一次，
否则会发生无法解决的冲突性问题
*/
func (r *Resources) Init() func(pk *packet.Packet) {
	*r = Resources{
		Command: commandRequestWithResponce{
			requestWithResponce: sync.Map{},
		},
		Inventory: inventoryContents{
			lockDown: sync.RWMutex{},
			datas:    make(map[uint32]map[uint8]protocol.ItemInstance),
		},
		ItemStackOperation: itemStackReuqestWithResponce{
			requestWithResponce: sync.Map{},
			currentRequestID:    1,
		},
		Container: container{
			lockDown:           sync.RWMutex{},
			containerOpenData:  nil,
			containerCloseData: nil,
			responded:          make(chan struct{}, 1),
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
	}
	// init struct
	return r.handlePacket
	// return
}
