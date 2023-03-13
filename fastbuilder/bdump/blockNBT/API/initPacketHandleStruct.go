package blockNBT_API

import (
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"

	"github.com/google/uuid"
)

var haveBeenDown bool = false // 用于标识是否已经在启动时初始化了 PacketHandleResult 结构体

// 初始化 PacketHandleResult 结构体中的各个参数。
// 此函数仅能在启动时由 PhoenixBuilder 统一调用，
// 且您不能创建一个新的 PacketHandleResult 结构体，因为在缺失包处理的情况下是无意义的。
func (p *PacketHandleResult) InitValue() {
	// 如果已经初始化了
	if haveBeenDown {
		return
	}
	// ----- commandDatas -----
	p.commandDatas = commandRequestWithResponce{
		commandRequest: struct {
			lockDown sync.RWMutex
			datas    map[uuid.UUID]*sync.Mutex
		}{
			lockDown: sync.RWMutex{},
			datas:    make(map[uuid.UUID]*sync.Mutex),
		},
		commandResponce: struct {
			lockDown sync.RWMutex
			datas    map[uuid.UUID]packet.CommandOutput
		}{
			lockDown: sync.RWMutex{},
			datas:    make(map[uuid.UUID]packet.CommandOutput),
		},
	}
	// ----- Inventory -----
	p.Inventory = inventoryContents{
		lockDown: sync.RWMutex{},
		datas:    make(map[uint32]map[uint8]protocol.ItemInstance),
	}
	// ----- ItemStackOperation -----
	p.ItemStackOperation = itemStackReuqestWithResponce{
		itemStackRequest: struct {
			lockDown sync.RWMutex
			datas    map[int32]*sync.Mutex
		}{
			lockDown: sync.RWMutex{},
			datas:    make(map[int32]*sync.Mutex),
		},
		itemStackResponce: struct {
			lockDown sync.RWMutex
			datas    map[int32]protocol.ItemStackResponse
		}{
			lockDown: sync.RWMutex{},
			datas:    make(map[int32]protocol.ItemStackResponse),
		},
		requestID: 1,
	}
	// ----- ContainerResources -----
	p.ContainerResources = container{
		containerOpen: struct {
			lockDown sync.RWMutex
			datas    packet.ContainerOpen
		}{
			lockDown: sync.RWMutex{},
			datas:    packet.ContainerOpen{},
		},
		containerClose: struct {
			lockDown sync.RWMutex
			datas    packet.ContainerClose
		}{
			lockDown: sync.RWMutex{},
			datas:    packet.ContainerClose{},
		},
		awaitChanges: sync.Mutex{},
		isUsing:      sync.Mutex{},
	}
	// 标记 PacketHandleResult 结构体的初始化状态为真
	haveBeenDown = true
}
