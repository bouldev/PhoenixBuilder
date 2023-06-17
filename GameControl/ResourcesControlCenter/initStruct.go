package ResourcesControlCenter

import (
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
)

// Resources 最多只能被初始化一次，因为资源在 PhoenixBuilder 中是唯一的
var hasInited bool = false

/*
初始化 Resources 结构体并返回一个函数用于更新资源。

此函数在每次启动 PhoenixBuilder 后至多调用一次，
重复的调用会导致程序惊慌，因为客户端的各项资源在同一时刻至多存在一个
*/
func (r *Resources) Init() func(pk *packet.Packet) {
	if !hasInited {
		hasInited = true
	} else {
		panic("Init: Attempts to obtain the client public resource multiple times")
	}
	// test if has been inited
	r.verified = true
	// verified
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
			responded:          make(chan bool, 1),
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
