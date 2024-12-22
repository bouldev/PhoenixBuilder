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
	"phoenixbuilder/minecraft/protocol/packet"
	"time"
)

// 用于在发送容器相关的数据包前执行，
// 便于后续调用 AwaitChangesAfterSendPacket 以阻塞程序的执行从而
// 达到等待租赁服响应容器操作的目的。
//
// 无论如何，即便不需要得到响应，也仍然需要使用此函数。
func (c *container) AwaitChangesBeforeSendingPacket() {
	c.responded = make(chan struct{}, 1)
}

/*
用于已经向租赁服提交容器操作后执行，
以等待租赁服响应容器的打开或关闭操作。
在调用此函数后，会持续等待直到租赁服响应这些请求。

如果租赁服在最长截止时间到来后依旧未对这些请求响应，
那么此函数将会返回值。
您可以通过 c.Getcontainer_opening_data() 或
c.Getcontainer_closing_data() 来验证容器是否
正确打开或关闭。
*/
func (c *container) AwaitChangesAfterSendingPacket() {
	select {
	case <-c.responded:
		return
	case <-time.After(ContainerOperationDeadLine):
		return
	}
}

// 向 c.responded 发送已响应的通知。
// 如果容器资源未被占用，则通知不会被发送。
// 当且仅当租赁服确认客户端的容器操作时，此函数才会被调用。
// 属于私有实现
func (c *container) respond_to_container_operation() {
	if c.GetOccupyStates() {
		c.responded <- struct{}{}
		close(c.responded)
	}
}

// 向 c.containerOpenData 写入容器开启数据 data ，属于私有实现
func (c *container) write_container_opening_data(data *packet.ContainerOpen) {
	c.lock_down.Lock()
	defer c.lock_down.Unlock()
	c.container_opening_data = data
}

// 取得当前已打开容器的数据。
// 如果容器未被打开或已被关闭，则会返回 nil 。
// 返回值虽然是一个地址，但它所指向的实际是一个副本
func (c *container) GetContainerOpeningData() *packet.ContainerOpen {
	c.lock_down.RLock()
	defer c.lock_down.RUnlock()
	// lock down
	if c.container_opening_data == nil {
		return nil
	} else {
		new := *c.container_opening_data
		return &new
	}
	// return
}

// 向 c.containerCloseData 写入容器关闭数据 data ，属于私有实现
func (c *container) write_container_closing_data(data *packet.ContainerClose) {
	c.lock_down.Lock()
	defer c.lock_down.Unlock()
	c.container_closing_data = data
}

// 取得上次关闭容器时租赁服的响应数据。
// 如果现在有容器已被打开或容器从未被关闭，则会返回 nil 。
// 返回值虽然是一个地址，但它所指向的实际是一个副本
func (c *container) GetContainerClosingData() *packet.ContainerClose {
	c.lock_down.RLock()
	defer c.lock_down.RUnlock()
	// lock down
	if c.container_closing_data == nil {
		return nil
	} else {
		new := *c.container_closing_data
		return &new
	}
	// return
}
