package GameInterface

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
	got, err := g.ParseTargetQueryingInfo(*resp.Respond)
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

// 在 pos 处以点击方块的形式放置朝向为 facing 的方块，
// hotBarSlot 指代要放置的方块在快捷栏的位置。
//
// 方块可以是潜影盒，也可以是旗帜
func (g *GameInterface) PlaceBlockWithFacing(
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
			return fmt.Errorf("PlaceBlockWithFacing: %v", resp.Error)
		}
		got, err := g.ParseTargetQueryingInfo(*resp.Respond)
		if err != nil {
			return fmt.Errorf("PlaceBlockWithFacing: %v", err)
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
			return fmt.Errorf("PlaceBlockWithFacing: %v", err)
		}
		defer g.RevertStructure(backupShulkerBoxUniqueId, pos)
	}
	// 可能目标方块并非生成在原本给定的坐标处，
	// 此时需要进行特殊处理
	err := g.SetBlockAsync(pos, "air", "[]")
	if err != nil {
		return fmt.Errorf("PlaceBlockWithFacing: %v", err)
	}
	err = g.SendSettingsCommand(teleportCommand, true)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithFacing: %v", err)
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
		return fmt.Errorf("PlaceBlockWithFacing: %v", err)
	}
	err = g.SetBlock(backupBlockPos, PlaceBlockBase, "[]")
	if err != nil {
		return fmt.Errorf("PlaceBlockWithFacing: %v", err)
	}
	// 清除目标方块处的方块、修正机器人的朝向、备份相关的方块，
	// 然后再在备份的方块处生成 PlaceBlockBase 所指代的方块。
	// 生成的 PlaceBlockBase 方块将被用于作为放置目标方块的依附方块
	err = g.ChangeSelectedHotbarSlot(hotBarSlot)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithFacing: %v", err)
	}
	err = g.PlaceBlock(
		UseItemOnBlocks{
			HotbarSlotID: hotBarSlot,
			BlockPos:     backupBlockPos,
			BlockName:    PlaceBlockBase,
			BlockStates:  map[string]interface{}{},
		},
		int32(facing),
	)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithFacing: %v", err)
	}
	// 更换手持物品栏为 hotBarSlot，
	// 然后点击 PlaceBlockBase 方块以放置目标方块
	err = g.RevertStructure(uniqueId, backupBlockPos)
	if err != nil {
		return fmt.Errorf("PlaceBlockWithFacing: %v", err)
	}
	// 将 PlaceBlockBase 处的方块恢复为原本方块
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
			return fmt.Errorf("PlaceBlockWithFacing: %v", err)
		}
		err = g.AwaitChangesGeneral()
		if err != nil {
			return fmt.Errorf("PlaceBlockWithFacing: %v", err)
		}
	}
	// 可能目标方块并非生成在原本给定的坐标处，
	// 此时需要进行特殊处理
	return nil
	// 返回值
}
