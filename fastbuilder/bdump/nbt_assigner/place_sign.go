package NBTAssigner

import (
	"fmt"
	GameInterface "phoenixbuilder/game_control/game_interface"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 我们不再检查用户提供的告示牌的 NBT 是否正确，
// 我们信任并且永远认为它们是正确且完整的
func (s *Sign) Decode() error {
	return nil
}

// 放置一个告示牌并写入告示牌数据
func (s *Sign) WriteData() error {
	api := s.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 初始化
	if s.BlockEntity.AdditionalData.FastMode {
		err := s.BlockEntity.Interface.SetBlockAsync(s.BlockEntity.AdditionalData.Position, s.BlockEntity.Block.Name, s.BlockEntity.AdditionalData.BlockStates)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		return nil
	}
	// 放置告示牌(快速导入模式下)
	{
		err := api.SendSettingsCommand(fmt.Sprintf("tp %d %d %d", s.BlockEntity.AdditionalData.Position[0], s.BlockEntity.AdditionalData.Position[1], s.BlockEntity.AdditionalData.Position[2]), true)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 传送机器人到告示牌所在的位置
		err = api.SetBlock(s.BlockEntity.AdditionalData.Position, s.BlockEntity.Block.Name, s.BlockEntity.AdditionalData.BlockStates)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 放置方块
	}
	// 放置告示牌
	err := api.ChangeSelectedHotbarSlot(0)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 切换手持物品栏到快捷栏 0
	err = api.ClickBlock(GameInterface.UseItemOnBlocks{
		HotbarSlotID: 0,
		BlockPos:     s.BlockEntity.AdditionalData.Position,
		BlockName:    s.BlockEntity.Block.Name,
		BlockStates:  s.BlockEntity.Block.States,
	})
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 打开告示牌
	err = api.WritePacket(&packet.BlockActorData{
		Position: s.BlockEntity.AdditionalData.Position,
		NBTData:  s.BlockEntity.Block.NBT,
	})
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 写入告示牌数据
	return nil
	// 返回值
}
