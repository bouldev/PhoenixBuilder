package ResourcesControl

import (
	"fmt"
	"phoenixbuilder/fastbuilder/sync_map"
	"phoenixbuilder/minecraft/protocol"
	"sort"
)

// 列出所有可访问库存的窗口 ID ，即 WindowID 。
// 返回的切片内的数字信息是无序的
func (i *inventoryContents) ListWindowID() []uint32 {
	i.lockDown.RLock()
	defer i.lockDown.RUnlock()
	// init
	ans := []uint32{}
	i.datas.Range(func(key uint32, value *sync_map.Map[uint8, protocol.ItemInstance]) bool {
		ans = append(ans, key)
		return true
	})
	// get window id list
	return ans
	// return
}

// 列出指定窗口 ID 所对应库存中的所有已记录槽位。
// 返回的切片内的数字信息将以升序排序。
//
// filter 是一个可选的过滤器，在其生效时，
// 将只返回满足条件的物品栏。
// 假设苹果和空气的 ItemRunTimeId 分别为 233 和 0 ，
// 则使用 []int32{233, 0} 作为过滤器时，将只会返回
// 槽位中为苹果和空气的物品栏编号。
//
// 如果不希望使用过滤器，那么请在此参数处填写 nil
func (i *inventoryContents) ListSlot(
	windowID uint32,
	filter *[]int32,
) ([]uint8, error) {
	i.lockDown.RLock()
	defer i.lockDown.RUnlock()
	// lock down resources
	got, ok := i.datas.Load(windowID)
	if !ok {
		return []uint8{}, fmt.Errorf("ListSlot: %v is not recorded in i.datas", windowID)
	}
	// if windowsID is not exist
	newFilter := map[int32]interface{}{}
	if filter != nil {
		for _, value := range *filter {
			newFilter[value] = true
		}
	}
	// init map for filter
	tmp := []int{}
	got.Range(func(key uint8, value protocol.ItemInstance) bool {
		if filter == nil || newFilter[value.Stack.ItemType.NetworkID] != nil {
			tmp = append(tmp, int(key))
		}
		return true
	})
	// get slots list
	sort.Stable(sort.IntSlice(tmp))
	// sort
	res := []byte{}
	for _, value := range tmp {
		res = append(res, byte(value))
	}
	return res, nil
	// return
}

// 获取 windowID 所对应的库存数据
func (i *inventoryContents) GetInventoryInfo(windowID uint32) (
	inventory map[uint8]protocol.ItemInstance,
	err error,
) {
	i.lockDown.RLock()
	defer i.lockDown.RUnlock()
	// init
	res, ok := i.datas.Load(windowID)
	if !ok {
		return nil, fmt.Errorf("GetInventoryInfo: %v is not recorded in i.datas", windowID)
	}
	// get data and check condition
	inventory = make(map[uint8]protocol.ItemInstance)
	res.Range(func(key uint8, value protocol.ItemInstance) bool {
		inventory[key] = value
		return true
	})
	// sync data to inventory
	return
	// return
}

// 从 windowID 库存中获取 slotLocation 槽位的物品数据
func (i *inventoryContents) GetItemStackInfo(windowID uint32, slotLocation uint8) (
	protocol.ItemInstance,
	error,
) {
	i.lockDown.RLock()
	defer i.lockDown.RUnlock()
	// init
	got, ok := i.datas.Load(windowID)
	if !ok {
		return protocol.ItemInstance{}, fmt.Errorf("GetItemStackInfo: %v is not recorded in i.datas", windowID)
	}
	// if windowsID is not exist
	ret, ok := got.Load(slotLocation)
	if !ok {
		return protocol.ItemInstance{}, fmt.Errorf("GetItemStackInfo: %v is not recorded in i.datas[%v]; i.datas[%v] = %#v", slotLocation, windowID, windowID, got)
	}
	// if slot is not exist
	return ret, nil
	// return
}

// 创建窗口 ID 为 windowID 的库存，如果库存不存在的话
func (i *inventoryContents) createNewInventory(windowID uint32) {
	i.lockDown.Lock()
	defer i.lockDown.Unlock()
	// init
	if _, ok := i.datas.Load(windowID); !ok {
		i.datas.Store(windowID, &sync_map.Map[uint8, protocol.ItemInstance]{})
	}
	// create new inventory
}

// 修改 windowID 库存中 slotLocation 槽位的物品数据，属于私有实现
func (i *inventoryContents) writeItemStackInfo(
	windowID uint32,
	slotLocation uint8,
	itemStackInfo protocol.ItemInstance,
) {
	i.createNewInventory(windowID)
	// create new inventory if needed
	i.lockDown.Lock()
	defer i.lockDown.Unlock()
	// lock down resources
	target_inventory, _ := i.datas.Load(windowID)
	target_inventory.Store(slotLocation, itemStackInfo)
	// write datas
}

// 删除 windowID 所对应的库存。
// 例如，当容器被关闭后，那么可以通过此函数删除此容器的库存数据，属于私有实现
func (i *inventoryContents) deleteInventory(windowID uint32) error {
	i.lockDown.Lock()
	defer i.lockDown.Unlock()
	// init
	_, ok := i.datas.Load(windowID)
	if !ok {
		return fmt.Errorf("deleteInventory: %v is not recorded in i.datas", windowID)
	}
	// if windowID is not exist
	i.datas.Delete(windowID)
	// remove inventory from i.datas
	return nil
	// return
}
