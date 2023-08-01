package GameInterface

import (
	"fmt"
	ResourcesControl "phoenixbuilder/game_control/resources_control"

	"github.com/google/uuid"
)

/*
在 pos 处尝试放置一个方块状态为 blockStates 的铁砧并附带承重方块。
考虑到给定的 pos 可能已经超出了客户端所在维度的高度限制，因此此函数将会进行自适应处理，
并在返回值 [3]int32 部分告知铁砧最终生成的位置。

由于承重方块会替换 pos 下方一格原本的方块，所以会使用 structure 命令备份一次。
结构的名称将对应 uuid_to_safe_string(uuid.UUID) 的字符串形式。
被备份结构包含 2 个方块，分别对应铁砧和承重方块原本的方块。

另，请使用 GameInterface.RevertStructure 来恢复铁砧和承重方块为原本方块
*/
func (g *GameInterface) GenerateNewAnvil(pos [3]int32, blockStates string) (
	uuid.UUID, [3]int32, error,
) {
	resp := g.SendWSCommandWithResponse(
		"querytarget @s",
		ResourcesControl.CommandRequestOptions{
			TimeOut: ResourcesControl.CommandRequestNoDeadLine,
		},
	)
	if resp.Error != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", resp.Error)
	}
	got, err := g.ParseTargetQueryingInfo(resp.Respond)
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	datas := got[0].Dimension
	// 取得客户端当前所在的维度
	switch datas {
	case OverWorldID:
		if pos[1] > OverWorld_MaxPosy {
			pos[1] = OverWorld_MaxPosy
		}
	case NetherID:
		if pos[1] > Nether_MaxPosy {
			pos[1] = Nether_MaxPosy
		}
	case EndID:
		if pos[1] > End_MaxPosy {
			pos[1] = End_MaxPosy
		}
	}
	// 如果放置坐标超出了最高限制
	switch datas {
	case OverWorldID:
		if pos[1]-1 < OverWorld_MinPosy {
			pos[1] = OverWorld_MinPosy + 1
		}
	case NetherID:
		if pos[1]-1 < Nether_MinPosy {
			pos[1] = Nether_MinPosy + 1
		}
	case EndID:
		if pos[1]-1 < End_MinPosy {
			pos[1] = End_MinPosy + 1
		}
	}
	// 如果放置坐标低于了最低限制
	uniqueId, err := g.BackupStructure(
		MCStructure{
			BeginX: pos[0],
			BeginY: pos[1] - 1,
			BeginZ: pos[2],
			SizeX:  1,
			SizeY:  2,
			SizeZ:  1,
		},
	)
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	// 备份相关的方块
	err = g.SendSettingsCommand(fmt.Sprintf("setblock %d %d %d %s", pos[0], pos[1]-1, pos[2], AnvilBase), true)
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	err = g.SendSettingsCommand(fmt.Sprintf("setblock %d %d %d anvil %s", pos[0], pos[1], pos[2], blockStates), true)
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	err = g.AwaitChangesGeneral()
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: Failed to generate a new anvil on %v; err = %#v", pos, err)
	}
	// 放置一个铁砧并附带一个承重方块
	return uniqueId, pos, nil
	// 返回值
}

// 在 pos 处以点击方块的形式放置朝向为 facing 的潜影盒。
// hotBarSlot 指代该潜影盒在快捷栏的位置。
// 我们将会使用该快捷栏的物品，
// 然后点击对应的方块以达到放置指定朝向的潜影盒的目的
func (g *GameInterface) PlaceShulkerBox(
	pos [3]int32,
	hotBarSlot uint8,
	facing uint8,
) error {
	var originPos [3]int32 = pos
	var teleportCommand string
	var backupBlockPos [3]int32
	// 初始化
	if facing == 0 || facing == 1 {
		resp := g.SendWSCommandWithResponse(
			"querytarget @s",
			ResourcesControl.CommandRequestOptions{
				TimeOut: ResourcesControl.CommandRequestNoDeadLine,
			},
		)
		if resp.Error != nil {
			return fmt.Errorf("PlaceShulkerBox: %v", resp.Error)
		}
		got, err := g.ParseTargetQueryingInfo(resp.Respond)
		if err != nil {
			return fmt.Errorf("PlaceShulkerBox: %v", err)
		}
		datas := got[0].Dimension
		// 取得客户端所在的维度
		switch datas {
		case OverWorldID:
			if pos[1]+1 > OverWorld_MaxPosy {
				pos[1] = OverWorld_MaxPosy - 1
			}
		case NetherID:
			if pos[1]+1 > Nether_MaxPosy {
				pos[1] = Nether_MaxPosy - 1
			}
		case EndID:
			if pos[1]+1 > End_MaxPosy {
				pos[1] = End_MaxPosy - 1
			}
		}
		// 如果放置坐标超出了最高限制
		switch datas {
		case OverWorldID:
			if pos[1]-1 < OverWorld_MinPosy {
				pos[1] = OverWorld_MinPosy + 1
			}
		case NetherID:
			if pos[1]-1 < Nether_MinPosy {
				pos[1] = Nether_MinPosy + 1
			}
		case EndID:
			if pos[1]-1 < End_MinPosy {
				pos[1] = End_MinPosy + 1
			}
		}
		// 如果放置坐标低于了最低限制
	}
	// 超高度的自适应处理
	backupBlockPos = pos
	switch facing {
	case 0:
		backupBlockPos[1] = backupBlockPos[1] + 1
		teleportCommand = "tp ~ ~ ~ 0.0 -90.0"
	case 1:
		backupBlockPos[1] = backupBlockPos[1] - 1
		teleportCommand = "tp ~ ~ ~ 0.0 90.0"
	case 2:
		backupBlockPos[2] = backupBlockPos[2] + 1
		teleportCommand = "tp ~ ~ ~ facing ~ ~ ~1.0"
	case 3:
		backupBlockPos[2] = backupBlockPos[2] - 1
		teleportCommand = "tp ~ ~ ~ facing ~ ~ ~-1.0"
	case 4:
		backupBlockPos[0] = backupBlockPos[0] + 1
		teleportCommand = "tp ~ ~ ~ facing ~1.0 ~ ~"
	case 5:
		backupBlockPos[0] = backupBlockPos[0] - 1
		teleportCommand = "tp ~ ~ ~ facing ~-1.0 ~ ~"
	}
	// 确定被备份方块的位置和用于修正机器人朝向的命令
	if pos != originPos {
		backupShulkerBoxUniqueId, err := g.BackupStructure(
			MCStructure{
				BeginX: pos[0],
				BeginY: pos[1],
				BeginZ: pos[2],
				SizeX:  1,
				SizeY:  1,
				SizeZ:  1,
			},
		)
		if err != nil {
			return fmt.Errorf("PlaceShulkerBox: %v", err)
		}
		defer g.RevertStructure(backupShulkerBoxUniqueId, pos)
	}
	// 可能潜影盒并非生成在原本给定的坐标处，此时需要进行特殊处理
	err := g.SetBlockAsync(pos, "air", "[]")
	if err != nil {
		return fmt.Errorf("PlaceShulkerBox: %v", err)
	}
	err = g.SendSettingsCommand(teleportCommand, true)
	if err != nil {
		return fmt.Errorf("PlaceShulkerBox: %v", err)
	}
	uniqueId, err := g.BackupStructure(
		MCStructure{
			BeginX: backupBlockPos[0],
			BeginY: backupBlockPos[1],
			BeginZ: backupBlockPos[2],
			SizeX:  1,
			SizeY:  1,
			SizeZ:  1,
		},
	)
	if err != nil {
		return fmt.Errorf("PlaceShulkerBox: %v", err)
	}
	err = g.SetBlock(backupBlockPos, "emerald_block", "[]")
	if err != nil {
		return fmt.Errorf("PlaceShulkerBox: %v", err)
	}
	// 清除潜影盒处的方块、修正机器人的朝向、备份相关的方块，
	// 然后再在备份的方块处生成一个绿宝石块。
	// 生成的绿宝石块将被用于作为放置潜影盒的依附方块。
	// SuperScript 最喜欢绿宝石块了！
	err = g.ChangeSelectedHotbarSlot(hotBarSlot)
	if err != nil {
		return fmt.Errorf("PlaceShulkerBox: %v", err)
	}
	err = g.PlaceBlock(
		UseItemOnBlocks{
			HotbarSlotID: hotBarSlot,
			BlockPos:     backupBlockPos,
			BlockName:    "minecraft:emerald_block",
			BlockStates:  map[string]interface{}{},
		},
		int32(facing),
	)
	if err != nil {
		return fmt.Errorf("PlaceShulkerBox: %v", err)
	}
	// 更换手持物品栏为 hotBarSlot 并点击绿宝石块以放置潜影盒。
	err = g.RevertStructure(uniqueId, backupBlockPos)
	if err != nil {
		return fmt.Errorf("PlaceShulkerBox: %v", err)
	}
	// 将绿宝石块处的方块恢复为原本方块
	if pos != originPos {
		request := fmt.Sprintf(
			"clone %d %d %d %d %d %d %d %d %d",
			pos[0],
			pos[1],
			pos[2],
			pos[0],
			pos[1],
			pos[2],
			originPos[0],
			originPos[1],
			originPos[2],
		)
		err = g.SendSettingsCommand(request, true)
		if err != nil {
			return fmt.Errorf("PlaceShulkerBox: %v", err)
		}
		err = g.AwaitChangesGeneral()
		if err != nil {
			return fmt.Errorf("PlaceShulkerBox: %v", err)
		}
	}
	// 可能潜影盒并非生成在原本给定的坐标处，此时需要进行特殊处理
	return nil
	// 返回值
}
