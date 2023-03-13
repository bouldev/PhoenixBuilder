package blockNBT_API

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
)

// ------------------------- commandRequestWithResponce -------------------------

// 测定 key 是否在 c.commandRequest.datas 中。如果存在，那么返回真，否则返回假
func (c *commandRequestWithResponce) testRequest(key uuid.UUID) bool {
	c.commandRequest.lockDown.RLock()
	defer c.commandRequest.lockDown.RUnlock()
	// init
	_, ok := c.commandRequest.datas[key]
	return ok
	// return
}

// 测定 key 是否在 c.commandResponce.datas 中。如果存在，那么返回真，否则返回假
func (c *commandRequestWithResponce) testResponce(key uuid.UUID) bool {
	c.commandResponce.lockDown.RLock()
	defer c.commandResponce.lockDown.RUnlock()
	// init
	_, ok := c.commandResponce.datas[key]
	return ok
	// return
}

// 将名为 key 的命令请求放入 c.commandRequest.datas 并占用(锁定)此请求对应的互斥锁
func (c *commandRequestWithResponce) writeRequest(key uuid.UUID) error {
	if c.testRequest(key) {
		return fmt.Errorf("writeRequest: %v is already exist in c.commandRequest.datas", key.String())
	}
	// if key is already exist
	c.commandRequest.lockDown.Lock()
	// lock down resources
	c.commandRequest.datas[key] = &sync.Mutex{}
	c.commandRequest.datas[key].Lock()
	// lock down command request
	c.commandRequest.lockDown.Unlock()
	// unlock resources
	return nil
	// return
}

// 将名为 key 的命令请求从 c.commandRequest.datas 中移除并释放此请求对应的互斥锁
func (c *commandRequestWithResponce) deleteRequest(key uuid.UUID) error {
	if !c.testRequest(key) {
		return fmt.Errorf("deleteRequest: %v is not recorded in c.commandRequest.datas", key.String())
	}
	// if key is not exist
	c.commandRequest.lockDown.Lock()
	// lock down resources
	tmp := c.commandRequest.datas[key]
	// get tmp of the current resources
	delete(c.commandRequest.datas, key)
	newMap := map[uuid.UUID]*sync.Mutex{}
	for k, value := range c.commandRequest.datas {
		newMap[k] = value
	}
	c.commandRequest.datas = newMap
	// remove the key and values from c.commandRequest.datas
	c.commandRequest.lockDown.Unlock()
	// unlock resources
	tmp.Unlock()
	// unlock command request
	return nil
	// return
}

// 将命令请求的返回值写入 c.commandResponce.datas 并释放 c.commandRequest.datas 中对应的互斥锁
func (c *commandRequestWithResponce) writeResponce(key uuid.UUID, resp packet.CommandOutput) error {
	c.commandResponce.lockDown.Lock()
	defer c.commandResponce.lockDown.Unlock()
	// init
	c.commandResponce.datas[key] = resp
	// send command responce
	err := c.deleteRequest(key)
	if err != nil {
		return fmt.Errorf("writeResponce: %v", err)
	}
	// remove command reuqest from c.commandRequest.datas
	return nil
	// return
}

// 从 c.commandResponce.datas 读取名为 key 的命令请求的返回值并将此返回值从 c.commandResponce.datas 移除
func (c *commandRequestWithResponce) loadResponceAndDelete(key uuid.UUID) (packet.CommandOutput, error) {
	if !c.testResponce(key) {
		return packet.CommandOutput{}, fmt.Errorf("loadResponceAndDelete: %v is not recorded in c.commandResponce.datas", key.String())
	}
	// if key is not exist
	c.commandResponce.lockDown.Lock()
	// lock down resources
	ans := c.commandResponce.datas[key]
	newMap := map[uuid.UUID]packet.CommandOutput{}
	for k, value := range c.commandResponce.datas {
		newMap[k] = value
	}
	c.commandResponce.datas = newMap
	// get responce and remove the key and values from c.commandResponce.datas
	c.commandResponce.lockDown.Unlock()
	// unlock resources
	return ans, nil
	// return
}

// 等待租赁服响应命令请求 key 。在调用此函数后，会持续阻塞直到此命令请求所对应的互斥锁被释放
func (c *commandRequestWithResponce) awaitResponce(key uuid.UUID) {
	if !c.testRequest(key) {
		return
	}
	// if key is not exist
	c.commandRequest.lockDown.RLock()
	// lock down resources
	tmp := c.commandRequest.datas[key]
	// get tmp of the current resources
	c.commandRequest.lockDown.RUnlock()
	// unlock resources
	tmp.Lock()
	tmp.Unlock()
	// await responce
}

// ------------------------- inventoryContents -------------------------

// 列出所有可访问库存的窗口 ID ，即 WindowID
func (i *inventoryContents) ListWindowID() []uint32 {
	i.lockDown.RLock()
	defer i.lockDown.RUnlock()
	// init
	ans := []uint32{}
	for key := range i.datas {
		ans = append(ans, key)
	}
	// get window id list
	return ans
	// return
}

// 列出指定窗口 ID 所对应库存中的所有已记录槽位
func (i *inventoryContents) ListSlot(windowID uint32) ([]uint8, error) {
	i.lockDown.RLock()
	defer i.lockDown.RUnlock()
	// init
	got, ok := i.datas[windowID]
	if !ok {
		return []uint8{}, fmt.Errorf("ListSlot: %v is not recorded in i.datas; i.datas = %#v", windowID, i.datas)
	}
	// if windowsID is not exist
	ans := []uint8{}
	for key := range got {
		ans = append(ans, key)
	}
	// get slots list
	return ans, nil
	// return
}

// 获取 windowID 所对应的库存数据
func (i *inventoryContents) GetInventoryInfo(windowID uint32) (map[uint8]protocol.ItemInstance, error) {
	i.lockDown.RLock()
	defer i.lockDown.RUnlock()
	// init
	ans, ok := i.datas[windowID]
	if !ok {
		return map[uint8]protocol.ItemInstance{}, fmt.Errorf("GetInventoryInfo: %v is not recorded in i.datas; i.datas = %#v", windowID, i.datas)
	}
	// if windowsID is not exist
	return ans, nil
	// return
}

// 从 windowID 库存中获取 slotLocation 槽位的物品数据
func (i *inventoryContents) GetItemStackInfo(windowID uint32, slotLocation uint8) (protocol.ItemInstance, error) {
	i.lockDown.RLock()
	defer i.lockDown.RUnlock()
	// init
	got, ok := i.datas[windowID]
	if !ok {
		return protocol.ItemInstance{}, fmt.Errorf("GetItemStackInfo: %v is not recorded in i.datas; i.datas = %#v", windowID, i.datas)
	}
	// if windowsID is not exist
	ans, ok := got[slotLocation]
	if !ok {
		return protocol.ItemInstance{}, fmt.Errorf("GetItemStackInfo: %v is not recorded in i.datas[%v]; i.datas[%v] = %#v", slotLocation, windowID, windowID, i.datas[windowID])
	}
	// if slot is not exist
	return ans, nil
	// return
}

// 修改 windowID 库存中 slotLocation 槽位的物品数据，属于私有实现
func (i *inventoryContents) writeItemStackInfo(windowID uint32, slotLocation uint8, itemStackInfo protocol.ItemInstance) {
	i.lockDown.Lock()
	defer i.lockDown.Unlock()
	// init
	if i.datas == nil {
		i.datas = make(map[uint32]map[uint8]protocol.ItemInstance)
	}
	if i.datas[windowID] == nil {
		i.datas[windowID] = make(map[uint8]protocol.ItemInstance)
	}
	// make
	i.datas[windowID][slotLocation] = itemStackInfo
	// write datas
}

// 删除 windowID 所对应的库存。例如，当容器被关闭后，那么可以通过此函数删除此容器的库存数据，属于私有实现
func (i *inventoryContents) deleteInventory(windowID uint32) error {
	i.lockDown.Lock()
	defer i.lockDown.Unlock()
	// init
	_, ok := i.datas[windowID]
	if !ok {
		return fmt.Errorf("deleteInventory: %v is not recorded in i.datas; i.datas = %#v", windowID, i.datas)
	}
	// if windowID is not exist
	delete(i.datas, windowID)
	newMap := map[uint32]map[uint8]protocol.ItemInstance{}
	for key, value := range i.datas {
		newMap[key] = value
	}
	i.datas = newMap
	// remove inventory from i.datas
	return nil
	// return
}

// ------------------------- itemStackReuqestWithResponce -------------------------

// 测定请求 ID 为 key 的物品操作请求是否在 i.itemStackRequest.datas 中。
// 如果存在，那么返回真，否则返回假
func (i *itemStackReuqestWithResponce) TestRequest(key int32) bool {
	i.itemStackRequest.lockDown.RLock()
	defer i.itemStackRequest.lockDown.RUnlock()
	// init
	_, ok := i.itemStackRequest.datas[key]
	return ok
	// return
}

// 测定请求 ID 为 key 的物品操作请求 key 是否在 i.itemStackResponce.datas 中。
// 如果存在，那么返回真，否则返回假
func (i *itemStackReuqestWithResponce) TestResponce(key int32) bool {
	i.itemStackResponce.lockDown.RLock()
	defer i.itemStackResponce.lockDown.RUnlock()
	// init
	_, ok := i.itemStackResponce.datas[key]
	return ok
	// return
}

// 将请求 ID 为 key 的物品操作请求放入 i.itemStackRequest.datas 并占用(锁定)此请求对应的互斥锁
func (i *itemStackReuqestWithResponce) WriteRequest(key int32) error {
	if i.TestRequest(key) {
		return fmt.Errorf("WriteRequest: %v is already exist in i.itemStackRequest.datas", key)
	}
	// if key is already exist
	i.itemStackRequest.lockDown.Lock()
	// lock down resources
	i.itemStackRequest.datas[key] = &sync.Mutex{}
	i.itemStackRequest.datas[key].Lock()
	// lock down item stack request
	i.itemStackRequest.lockDown.Unlock()
	// unlock resources
	return nil
	// return
}

// 将请求 ID 为 key 的物品操作请求从 i.itemStackRequest.datas 中移除并释放此请求对应的互斥锁
func (i *itemStackReuqestWithResponce) DeleteRequest(key int32) error {
	if !i.TestRequest(key) {
		return fmt.Errorf("DeleteRequest: %v is not recorded in i.itemStackRequest.datas", key)
	}
	// if key is not exist
	i.itemStackRequest.lockDown.Lock()
	// lock down resources
	tmp := i.itemStackRequest.datas[key]
	// get tmp of the current resources
	delete(i.itemStackRequest.datas, key)
	newMap := map[int32]*sync.Mutex{}
	for k, value := range i.itemStackRequest.datas {
		newMap[k] = value
	}
	i.itemStackRequest.datas = newMap
	// remove the key and values from i.itemStackRequest.datas
	i.itemStackRequest.lockDown.Unlock()
	// unlock resources
	tmp.Unlock()
	// unlock item stack request
	return nil
	// return
}

// 将请求 ID 为 key 的物品操作请求的返回值写入 i.itemStackResponce.datas
// 并释放 i.itemStackRequest.datas 中对应的互斥锁，属于私有实现
func (i *itemStackReuqestWithResponce) writeResponce(key int32, resp protocol.ItemStackResponse) error {
	i.itemStackResponce.lockDown.Lock()
	defer i.itemStackResponce.lockDown.Unlock()
	// init
	i.itemStackResponce.datas[key] = resp
	// send item stack responce
	err := i.DeleteRequest(key)
	if err != nil {
		return fmt.Errorf("writeResponce: %v", err)
	}
	// remove item stack reuqest from i.itemStackRequest.datas
	return nil
	// return
}

// 从 i.itemStackResponce.datas 读取请求 ID 为 key 的物品操作请求的返回值
// 并将此返回值从 i.itemStackResponce.datas 移除
func (i *itemStackReuqestWithResponce) LoadResponceAndDelete(key int32) (protocol.ItemStackResponse, error) {
	if !i.TestResponce(key) {
		return protocol.ItemStackResponse{}, fmt.Errorf("LoadResponceAndDelete: %v is not recorded in i.itemStackResponce.datas", key)
	}
	// if key is not exist
	i.itemStackResponce.lockDown.Lock()
	// lock down resources
	ans := i.itemStackResponce.datas[key]
	newMap := map[int32]protocol.ItemStackResponse{}
	for k, value := range i.itemStackResponce.datas {
		newMap[k] = value
	}
	i.itemStackResponce.datas = newMap
	// get responce and remove the key and values from i.itemStackResponce.datas
	i.itemStackResponce.lockDown.Unlock()
	// unlock resources
	return ans, nil
	// return
}

// 等待租赁服响应请求 ID 为 key 的物品操作请求。
// 在调用此函数后，会持续阻塞直到此物品操作请求所对应的互斥锁被释放
func (i *itemStackReuqestWithResponce) AwaitResponce(key int32) {
	if !i.TestRequest(key) {
		return
	}
	// if key is not exist
	i.itemStackRequest.lockDown.RLock()
	// lock down resources
	tmp := i.itemStackRequest.datas[key]
	// get tmp of the current resources
	i.itemStackRequest.lockDown.RUnlock()
	// unlock resources
	tmp.Lock()
	tmp.Unlock()
	// await responce
}

// 以原子操作获取上一次的请求 ID ，也就是 RequestID 。
// 如果从未进行过物品操作，则将会返回 1
func (i *itemStackReuqestWithResponce) GetCurrentRequestID() int32 {
	return atomic.LoadInt32(&i.requestID)
}

// 以原子操作获取一个唯一的请求 ID ，也就是 RequestID
func (i *itemStackReuqestWithResponce) GetNewRequestID() int32 {
	return atomic.AddInt32(&i.requestID, -2)
}

// ------------------------- container -------------------------

/*
占用客户端的容器资源。
当 tryMode 为真时，将尝试占用资源并返回占用结果，此对应返回值 bool 部分。
若 tryMode 为假，则返回值 bool 部分永远为真。
无论 tryMode 的真假如何，当且仅当函数调用者成功占用资源时，
才会在 *sync.Mutex 部分返回互斥锁，否则此参数返回 nil 。
如果后续要释放此资源，那么请直接调用返回的互斥锁的 Unlock 函数
*/
func (c *container) Occupy(tryMode bool) (bool, *sync.Mutex) {
	if tryMode {
		success := c.isUsing.TryLock()
		if success {
			return true, &c.isUsing
		} else {
			return false, nil
		}
	}
	// if is try mode
	c.isUsing.Lock()
	// lock down resources
	return true, &c.isUsing
	// return
}

// 强制释放容器资源，当且仅当租赁服强制关闭容器时会被使用，属于私有实现
func (c *container) release() {
	c.isUsing.TryLock()
	c.isUsing.Unlock()
}

// 用于在 打开/关闭 容器前执行，便于后续调用 AwaitResponceAfterSendPacket 以阻塞程序的执行从而
// 达到等待租赁服响应容器操作的目的
func (c *container) AwaitResponceBeforeSendPacket() {
	c.awaitChanges.Lock()
}

// 等待租赁服响应容器的打开或关闭操作。在调用此函数后，会持续阻塞直到相关操作所对应的互斥锁被释放
func (c *container) AwaitResponceAfterSendPacket() {
	c.awaitChanges.Lock()
	c.awaitChanges.Unlock()
}

// 释放 c.awaitChanges 中关于容器操作的互斥锁。如果互斥锁未被锁定，程序也仍不会发生惊慌。
// 当且仅当租赁服确认客户端的容器操作时，此函数才会被调用。
// 属于私有实现
func (c *container) releaseAwaitGoRoutine() {
	c.awaitChanges.TryLock()
	c.awaitChanges.Unlock()
}

// 将 datas 写入 c.containerOpen.datas ，属于私有实现
func (c *container) writeContainerOpenDatas(datas packet.ContainerOpen) {
	c.containerOpen.lockDown.Lock()
	defer c.containerOpen.lockDown.Unlock()
	// init
	c.containerOpen.datas = datas
	// set values
}

// 取得当前已打开容器的数据。如果容器未被打开或已被关闭，则会返回一个刚被初始化的结构体
func (c *container) GetContainerOpenDatas() packet.ContainerOpen {
	c.containerOpen.lockDown.RLock()
	defer c.containerOpen.lockDown.RUnlock()
	// init
	return c.containerOpen.datas
	// return
}

// 将 datas 写入 c.containerClose.datas ，属于私有实现
func (c *container) writeContainerCloseDatas(datas packet.ContainerClose) {
	c.containerClose.lockDown.Lock()
	defer c.containerClose.lockDown.Unlock()
	// init
	c.containerClose.datas = datas
	// set values
}

// 取得上次关闭容器时租赁服的响应数据。如果现在有容器已被打开或容器从未被关闭，则会返回一个刚被初始化的结构体
func (c *container) GetContainerCloseDatas() packet.ContainerClose {
	c.containerClose.lockDown.RLock()
	defer c.containerClose.lockDown.RUnlock()
	// init
	return c.containerClose.datas
	// return
}

// ------------------------- end -------------------------
