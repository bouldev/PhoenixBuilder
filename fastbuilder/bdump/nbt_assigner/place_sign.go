package NBTAssigner

import (
	"fmt"
	GameInterface "phoenixbuilder/game_control/game_interface"
	"phoenixbuilder/minecraft/protocol/packet"

	"github.com/google/uuid"
)

// 我们不再检查用户提供的告示牌的 NBT 是否正确，
// 我们信任并且永远认为它们是正确且完整的
func (s *Sign) Decode() error {
	return nil
}

// 放置一个告示牌并写入告示牌数据
func (s *Sign) WriteData() error {
	var uniqueID_1 uuid.UUID
	var uniqueID_2 uuid.UUID
	// 初始化变量
	if s.BlockEntity.AdditionalData.FastMode {
		err := s.BlockEntity.Interface.SetBlockAsync(s.BlockEntity.AdditionalData.Position, s.BlockEntity.Block.Name, s.BlockEntity.AdditionalData.BlockStates)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		return nil
	}
	gameInterface := s.BlockEntity.Interface.(*GameInterface.GameInterface)
	// 放置告示牌(快速导入模式下)
	{
		err := gameInterface.SendSettingsCommand(fmt.Sprintf("tp %d %d %d", s.BlockEntity.AdditionalData.Position[0], s.BlockEntity.AdditionalData.Position[1], s.BlockEntity.AdditionalData.Position[2]), true)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 传送机器人到告示牌所在的位置
		err = gameInterface.SendSettingsCommand(
			fmt.Sprintf(
				"setblock %d %d %d air",
				s.BlockEntity.AdditionalData.Position[0],
				s.BlockEntity.AdditionalData.Position[1],
				s.BlockEntity.AdditionalData.Position[2],
			),
			true,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 清除当前告示牌处的方块。
		// 如果不这么做且原本该处的方块是告示牌的话，
		// 那么 NBT 数据将会注入失败
		err = gameInterface.SendSettingsCommand(
			"replaceitem entity @s slot.hotbar 0 oak_sign",
			true,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 获取一个告示牌到快捷栏 0
		err = gameInterface.ChangeSelectedHotbarSlot(0)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 切换手持物品栏到快捷栏 0
		uniqueID_1, err = gameInterface.BackupStructure(
			GameInterface.MCStructure{
				BeginX: s.BlockEntity.AdditionalData.Position[0] + 1,
				BeginY: s.BlockEntity.AdditionalData.Position[1],
				BeginZ: s.BlockEntity.AdditionalData.Position[2],
				SizeX:  1,
				SizeY:  1,
				SizeZ:  1,
			},
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		/*
			我们会在告示牌的 (~1, ~, ~) 处生成一个玻璃，
			然后点击这个玻璃并指定点击的面是 4 以将手中的告示牌放上去。

			这样，我们就可以取得反作弊的认同，
			然后我们就可以向告示牌注入 NBT 数据了。

			但在生成玻璃前，我们需要备份这个玻璃原本的方块以方便之后恢复它
		*/
		err = gameInterface.SendSettingsCommand(
			fmt.Sprintf(
				"setblock %d %d %d glass",
				s.BlockEntity.AdditionalData.Position[0]+1,
				s.BlockEntity.AdditionalData.Position[1],
				s.BlockEntity.AdditionalData.Position[2],
			),
			true,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		err = gameInterface.AwaitChangesGeneral()
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 生成上文提到的玻璃。
		// TODO: 优化上方这段代码
		err = gameInterface.PlaceBlock(
			GameInterface.UseItemOnBlocks{
				HotbarSlotID: 0,
				BlockPos: [3]int32{
					s.BlockEntity.AdditionalData.Position[0] + 1,
					s.BlockEntity.AdditionalData.Position[1],
					s.BlockEntity.AdditionalData.Position[2],
				},
				BlockName:   "minecraft:glass",
				BlockStates: map[string]interface{}{},
			},
			4,
		)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 在玻璃上放置手中的告示牌
		err = gameInterface.SetBlockAsync(s.BlockEntity.AdditionalData.Position, s.BlockEntity.Block.Name, s.BlockEntity.AdditionalData.BlockStates)
		if err != nil {
			return fmt.Errorf("WriteData: %v", err)
		}
		// 现在玻璃上有了一个告示牌，这是我们刚刚放上去的，
		// 但这个告示牌的种类是 oak_sign ，且朝向固定，
		// 因此现在我们需要覆写这个告示牌的种类及朝向为正确的形式。
		// 经过测试，覆写操作不会导致 NBT 数据无法注入
	}
	// 放置告示牌
	err := gameInterface.WritePacket(&packet.BlockActorData{
		Position: s.BlockEntity.AdditionalData.Position,
		NBTData:  s.BlockEntity.Block.NBT,
	})
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 写入告示牌数据
	uniqueID_2, err = gameInterface.BackupStructure(GameInterface.MCStructure{
		BeginX: s.BlockEntity.AdditionalData.Position[0],
		BeginY: s.BlockEntity.AdditionalData.Position[1],
		BeginZ: s.BlockEntity.AdditionalData.Position[2],
		SizeX:  1,
		SizeY:  1,
		SizeZ:  1,
	})
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	/*
		备份告示牌处的方块。

		稍后我们会恢复上文提到的玻璃处的方块为原本方块，
		而此方块被恢复后，游戏会按照特性刷新它附近的方块，
		也就是告示牌方块。

		但我们无法保证刷新后，我们导入的告示牌仍然可以稳定存在，
		因为它可能会因为缺少依附方块而掉落。

		因此，我们现在备份一次告示牌，然后再恢复玻璃处的方块，
		然后再强行生成一次告示牌本身。

		注：这个解法并不优雅，而且会浪费时间，
		但它可以显著提高告示牌的存活概率，
		而且用户不希望为了告示牌而再导入一次 BDX 文件。

		TODO: 在某天推迟部分方块的导入顺序，
		使得告示牌这类依附型方块在最后再被导入
	*/
	err = gameInterface.RevertStructure(
		uniqueID_1,
		GameInterface.BlockPos{
			s.BlockEntity.AdditionalData.Position[0] + 1,
			s.BlockEntity.AdditionalData.Position[1],
			s.BlockEntity.AdditionalData.Position[2],
		},
	)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 将上文提到的玻璃处的方块恢复为原本的方块
	gameInterface.RevertStructure(
		uniqueID_2,
		GameInterface.BlockPos{
			s.BlockEntity.AdditionalData.Position[0],
			s.BlockEntity.AdditionalData.Position[1],
			s.BlockEntity.AdditionalData.Position[2],
		},
	)
	// 再强行生成一次告示牌本身以抑制其可能发生的掉落
	return nil
	// 返回值
}
