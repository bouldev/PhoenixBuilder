package GameInterface

import (
	"fmt"
	"github.com/google/uuid"
)

/*
在 pos 处尝试放置一个方块状态为 blockStates 的铁砧并附带承重方块。
考虑到给定的 pos 可能已经超出了客户端所在维度的高度限制，因此此函数将会进行自适应处理，
并在返回值 [3]int32 部分告知铁砧生成的最终坐标。

由于承重方块会替换 pos 下方一格原本的方块，所以会使用 structure 命令备份一次。
结构的名称将对应返回值 uuid.UUID 参数的字符串形式。
被备份结构包含 2 个方块，分别对应铁砧和承重方块原本的方块。

另，我们推荐您使用 GlobalAPI.RevertStructure 来恢复铁砧和承重方块为原本方块
*/
func (g *GameInterface) GenerateNewAnvil(pos [3]int32, blockStates string) (
	uuid.UUID, [3]int32, error,
) {
	resp, err := g.SendWSCommandWithResponse("querytarget @s")
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	got, err := g.ParseTargetQueryingInfo(resp)
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
	err = g.SendSettingsCommand(fmt.Sprintf("setblock %d %d %d %v", pos[0], pos[1]-1, pos[2], AnvilBase), true)
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	resp, err = g.SendWSCommandWithResponse(fmt.Sprintf("setblock %d %d %d anvil %v", pos[0], pos[1], pos[2], blockStates))
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	if resp.SuccessCount <= 0 && resp.OutputMessages[0].Message != "commands.setblock.noChange" {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: Failed to generate a new anvil on %v; resp = %#v", pos, resp)
	}
	// 放置一个铁砧并附带一个承重方块
	return uniqueId, pos, nil
	// 返回值
}
