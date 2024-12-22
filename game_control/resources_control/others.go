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
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// ------------------------- currentTick -------------------------

// 提交请求 ID 为 key 的请求用于获取当前的游戏刻
func (o *others) WriteGameTickRequest(key uuid.UUID) error {
	_, exist := o.current_tick_request_with_resp.Load(key)
	if exist {
		return fmt.Errorf("WriteGameTickRequest: %v has already existed", key.String())
	}
	// if key has already exist
	o.current_tick_request_with_resp.Store(key, make(chan int64, 1))
	return nil
	// return
}

// 根据租赁服返回的 packet.TickSync(resp) 包，
// 向所有请求过 获取当前游戏刻 的请求写入此响应体的
// ServerReceptionTimestamp 字段
func (o *others) write_tick_sync_resp(resp packet.TickSync) error {
	var err error = nil
	o.current_tick_request_with_resp.Range(func(key uuid.UUID, value chan int64) bool {
		value <- resp.ServerReceptionTimestamp
		close(value)
		return true
	})
	// write responce for all the request
	return err
	// return
}

// 读取请求 ID 为 key 的 获取当前游戏刻 的请求所对应返回值，
// 同时移除该请求
func (o *others) LoadTickSyncResponse(
	key uuid.UUID,
) (int64, error) {
	value, exist := o.current_tick_request_with_resp.Load(key)
	if !exist {
		return 0, fmt.Errorf("LoadTickSyncResponse: %v is not recorded", key.String())
	}
	// if key is not exist
	res := <-value
	o.current_tick_request_with_resp.Delete(key)
	return res, nil
	// return
}
