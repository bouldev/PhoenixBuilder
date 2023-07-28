package GameInterface

import (
	"fmt"
	ResourcesControl "phoenixbuilder/game_control/resources_control"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 用于获取当前的游戏刻。
// 此操作不会被立即完成，
// 因为它需要请求一个数据包
func (g *GameInterface) GetCurrentTick() (int64, error) {
	uniqueId := ResourcesControl.GenerateUUID()
	// get a new uuid
	err := g.Resources.Others.WriteCurrentTickRequest(uniqueId)
	if err != nil {
		return 0, fmt.Errorf("GetCurrentTick: %v", err)
	}
	// write request
	err = g.WritePacket(&packet.TickSync{
		ClientRequestTimestamp:   0,
		ServerReceptionTimestamp: 0,
	})
	if err != nil {
		return 0, fmt.Errorf("GetCurrentTick: %v", err)
	}
	// send packet
	ans, err := g.Resources.Others.Load_TickSync_Packet_Responce_and_Delete_Request(uniqueId)
	if err != nil {
		return 0, fmt.Errorf("GetCurrentTick: %v", err)
	}
	return ans, nil
	// load responce and return
}
