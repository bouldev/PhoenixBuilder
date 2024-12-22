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
	"fmt"
	"phoenixbuilder/minecraft/protocol"
	"sync/atomic"

	"github.com/pterm/pterm"
)

// 提交物品请求 ID 为 key 的物品操作。
// how_to_change 指代相应槽位的变动结果，这用于更新本地库存数据
func (i *item_stack_request_with_response) WriteRequest(
	key int32,
	how_to_change map[ContainerID]StackRequestContainerInfo,
) error {
	_, exist := i.request_with_response.Load(key)
	if exist {
		return fmt.Errorf("WriteRequest: %v has already existed", key)
	}
	// if key has already exist
	i.request_with_response.Store(
		key,
		singleitem_stack_request_with_response{
			resp:          make(chan protocol.ItemStackResponse, 1),
			how_to_change: how_to_change,
		},
	)
	// write request
	return nil
	// return
}

// 向请求 ID 为 key 的物品操作写入返回值 resp 。
// 当 key 不存在时，将会抛出程序 panic 。
// 属于私有实现。
func (i *item_stack_request_with_response) write_response(
	key int32,
	resp protocol.ItemStackResponse,
) {
	value, exist := i.request_with_response.Load(key)
	if !exist {
		panic("write_response: Attempt to send packet.ItemStackRequest without using ResourcesControlCenter")
	}
	// if key is not exist
	value.resp <- resp
	// send response
}

// 读取请求 ID 为 key 的物品操作请求的返回值，
// 同时移除此物品操作请求
func (i *item_stack_request_with_response) LoadResponseAndDelete(key int32) (protocol.ItemStackResponse, error) {
	value, exist := i.request_with_response.Load(key)
	if !exist {
		return protocol.ItemStackResponse{}, fmt.Errorf("LoadResponseAndDelete: %v is not recorded", key)
	}
	// if key is not exist
	ret := <-value.resp
	i.request_with_response.Delete(key)
	return ret, nil
	// return
}

// 以原子操作获取上一次的请求 ID ，即 RequestID 。
// 如果从未进行过物品操作，则将会返回 1
func (i *item_stack_request_with_response) GetCurrentRequestID() int32 {
	return atomic.LoadInt32(&i.current_request_id)
}

// 以原子操作获取一个新的请求 ID ，即 RequestID
func (i *item_stack_request_with_response) GetNewRequestID() int32 {
	return atomic.AddInt32(&i.current_request_id, -2)
}

// 利用 newItemName 更新 item 中存储的物品名称信息。
// 如果传入的 newItemName 为空字符串，则将会从 item 中移除物品名称信息
func (i *item_stack_request_with_response) SetItemName(
	item *protocol.ItemInstance,
	newItemName string,
) error {
	nbt := item.Stack.NBTData
	// get nbt datas
	defer func() {
		item.Stack.NBTData = nbt
	}()
	// while exit
	deleteDisplay := func() {
		delete(nbt, "display")
		newMap := map[string]interface{}{}
		for key, value := range nbt {
			newMap[key] = value
		}
		nbt = newMap
	}
	deleteName := func() {
		delete(nbt["display"].(map[string]interface{}), "Name")
		newMap := map[string]interface{}{}
		for key, value := range nbt["display"].(map[string]interface{}) {
			newMap[key] = value
		}
		nbt["display"] = newMap
	}
	// init func
	if newItemName != "" {
		_, ok := nbt["display"]
		if !ok {
			nbt["display"] = map[string]interface{}{}
		}
		_, normal := nbt["display"].(map[string]interface{})
		if !normal {
			return fmt.Errorf("SetItemName: Failed to convert nbt[\"display\"] into map[string]interface{}; nbt = %#v", nbt)
		}
		// display
		nbt["display"].(map[string]interface{})["Name"] = newItemName
		// name
		return nil
		// return
	} else {
		_, ok := nbt["display"]
		if !ok {
			return nil
		}
		_, normal := nbt["display"].(map[string]interface{})
		if !normal {
			return fmt.Errorf("SetItemName: Failed to convert nbt[\"display\"] into map[string]interface{}; nbt = %#v", nbt)
		}
		// display
		_, ok = nbt["display"].(map[string]interface{})["Name"]
		if ok {
			deleteName()
		}
		if len(nbt["display"].(map[string]interface{})) <= 0 {
			deleteDisplay()
		}
		// name
		return nil
		// return
	}
	// update names
}

// 根据 newItem 中预期的新数据和租赁服返回的 resp ，
// 返回完整的新物品数据。
func (i *item_stack_request_with_response) GetNewItemData(
	newItem protocol.ItemInstance,
	resp protocol.StackResponseSlotInfo,
) (protocol.ItemInstance, error) {
	newItem.Stack.Count = uint16(resp.Count)
	newItem.StackNetworkID = resp.StackNetworkID
	/*
		newItem.Stack.MetadataValue = uint32(resp.DurabilityCorrection) [This line of code has not been tested for validity]
	*/
	return newItem, nil
}

/*
根据租赁服返回的 resp 字段更新对应库存中对应槽位的物品数据。
inventory 必须是一个指针，它指向了客户端库存数据在内存中存放的地址。
此函数属于私有实现。

由于返回的 resp 字段不完整，因此此函数内部使用 i.LoadRequest(resp.RequestID)
来加载原有的请求数据，并访问其中描述的物品变动的预期结果，
然后依此字段和 resp 字段更新本地库存数据。
*/
func (i *item_stack_request_with_response) update_item_data(
	resp protocol.ItemStackResponse,
	inventory *inventory_contents,
) error {
	request, exist := i.request_with_response.Load(resp.RequestID)
	if !exist {
		panic("update_item_data: Attempt to send packet.ItemStackRequest without using ResourcesControlCenter")
	}
	// 加载物品操作请求
	for _, value := range resp.ContainerInfo {
		if request.how_to_change == nil {
			panic("update_item_data: Results of item changes are not provided(packet.ItemStackRequest related)")
		}
		currentRequest, ok := request.how_to_change[ContainerID(value.ContainerID)]
		if !ok {
			pterm.Warning.Printf(
				"update_item_data: request_got.how_to_change[%d] is not provided(packet.ItemStackRequest related); request_got.how_to_change = %#v; value = %#v\n",
				ContainerID(value.ContainerID),
				request.how_to_change,
				value,
			)
			return nil
		}
		// 数据检查
		for _, val := range value.SlotInfo {
			expectNewItem, ok := currentRequest.ChangeResult[val.Slot]
			if !ok {
				pterm.Warning.Printf(
					"update_item_data: currentRequest.ChangeResult[%d] is not provided(packet.ItemStackRequest related); currentRequest.ChangeResult = %#v; val = %#v\n",
					val.Slot,
					currentRequest.ChangeResult,
					val,
				)
				continue
			}
			// 数据检查
			newItem, err := i.GetNewItemData(
				expectNewItem,
				val,
			)
			if err != nil {
				panic(fmt.Sprintf("update_item_data: Failed to get new item data; currentRequest.ChangeResult[val.Slot] = %#v, val = %#v", currentRequest.ChangeResult[val.Slot], val))
			}
			// 取得物品的新数据
			inventory.write_item_stack_info(currentRequest.WindowID, val.Slot, newItem)
			// 将物品的新数据写入到本地库存中
		}
		// 更新本地库存中一个或多个物品的数据
	}
	// 设置物品数据
	return nil
	// 返回值
}
