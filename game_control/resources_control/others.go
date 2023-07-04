package ResourcesControl

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// ------------------------- currentTick -------------------------

// 提交请求 ID 为 key 的请求用于获取当前的游戏刻
func (o *others) WriteCurrentTickRequest(key uuid.UUID) error {
	_, exist := o.currentTickRequestWithResp.Load(key)
	if exist {
		return fmt.Errorf("WriteCurrentTickRequest: %v has already existed", key.String())
	}
	// if key has already exist
	o.currentTickRequestWithResp.Store(key, make(chan int64, 1))
	return nil
	// return
}

// 根据租赁服返回的 packet.TickSync(resp) 包，
// 向所有请求过 获取当前游戏刻 的请求写入此响应体的
// ServerReceptionTimestamp 字段
func (o *others) writeTickSyncPacketResponse(resp packet.TickSync) error {
	var err error = nil
	o.currentTickRequestWithResp.Range(func(key, value any) bool {
		chanGet, normal := value.(chan int64)
		if !normal {
			err = fmt.Errorf("writeTickSyncPacketResponse: Failed to convert value into (chan int64); key(RequestID) = %#v, value = %#v", key, value)
			return false
		}
		// convert data
		chanGet <- resp.ServerReceptionTimestamp
		close(chanGet)
		return true
		// return
	})
	// write responce for all the request
	return err
	// return
}

// 读取请求 ID 为 key 的 获取当前游戏刻 的请求所对应返回值，
// 同时移除该请求
func (o *others) Load_TickSync_Packet_Responce_and_Delete_Request(
	key uuid.UUID,
) (int64, error) {
	value, exist := o.currentTickRequestWithResp.Load(key)
	if !exist {
		return 0, fmt.Errorf("Load_TickSync_Packet_Responce_and_Delete_Request: %v is not recorded", key.String())
	}
	// if key is not exist
	chanGet, normal := value.(chan int64)
	if !normal {
		return 0, fmt.Errorf("Load_TickSync_Packet_Responce_and_Delete_Request: Failed to convert value into (chan int64); value = %#v", value)
	}
	// convert data
	res := <-chanGet
	o.currentTickRequestWithResp.Delete(key)
	return res, nil
	// return
}

// ------------------------- END -------------------------
