package GameInterface

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
	"time"
)

// 获取 pos 处的方块到物品栏。
// 返回的布尔值代表请求是否成功；
// 返回的 uint8 代表该方块最终生成在快捷栏的位置
func (g *GameInterface) PickBlock(
	pos [3]int32,
	assignNBTData bool,
) (bool, uint8, error) {
	var selectedHotBarSlot int8 = -1
	// 初始化
	for i := 0; i < BlockPickRequestReTryMaximumCounts; i++ {
		err := g.ChangeSelectedHotbarSlot(5)
		if err != nil {
			return false, 0, fmt.Errorf("PickBlock: %v", err)
		}
		// 将物品栏切换到 5
		listener, packets := g.Resources.Listener.CreateNewListen([]uint32{packet.IDPlayerHotBar}, 1)
		// 注册一个用于监听 packet.IDPlayerHotBar 的数据包监听器
		g.WritePacket(&packet.BlockPickRequest{
			Position:    pos,
			AddBlockNBT: assignNBTData,
			HotBarSlot:  5,
		})
		// 发送 Pick Block 请求
		select {
		case pk := <-packets:
			selectedHotBarSlot = int8(pk.(*packet.PlayerHotBar).SelectedHotBarSlot)
		case <-time.After(BlockPickRequestDeadLine):
		}
		// 确定方块是被 Pick 到了哪个物品栏
		err = g.Resources.Listener.StopAndDestroy(listener)
		if err != nil {
			return false, 0, fmt.Errorf("PickBlock: %v", err)
		}
		// 终止并关闭数据包监听器
		if selectedHotBarSlot == -1 {
			continue
		} else {
			return true, uint8(selectedHotBarSlot), nil
		}
		// 如果当次请求超时，则重试，否则直接返回值。
		// 最多尝试(总次数) BlockPickRequestReTryMaximumCounts 次
	}
	// 发送 Pick Block 请求并确定方块是被 Pick 到了哪个物品栏
	return false, 0, nil
	// 返回值
}
