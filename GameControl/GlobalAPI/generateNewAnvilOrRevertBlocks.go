package GlobalAPI

import (
	"fmt"

	"github.com/google/uuid"
)

// 描述各个维度可放置方块的最高高度
const (
	OverWorld_MaxPosy = int32(319) // 主世界
	Nether_MaxPosy    = int32(127) // 下界
	End_MaxPosy       = int32(255) // 末地
)

// 描述各个维度可放置方块的最低高度
const (
	OverWorld_MinPosy = int32(-64) // 主世界
	Nether_MinPosy    = int32(0)   // 下界
	End_MinPosy                    // 末地
)

const BlockUnderAnvil string = "glass" // 用作铁砧的承重方块

/*
在 pos 处尝试放置一个方块状态为 blockStates 的铁砧并附带承重方块。
考虑到给定的 pos 可能已经超出了客户端所在维度的高度限制，因此此函数将会进行自适应处理，
并在返回值 [3]int32 部分告知铁砧生成的最终坐标。

由于承重方块会替换 pos 下方一格原本的方块，所以会使用 structure 命令备份一次。
结构的名称将对应返回值 uuid.UUID 参数的字符串形式。
被备份结构包含 2 个方块，分别对应铁砧和承重方块原本的方块。

其他：
我们推荐您使用 GlobalAPI.RevertBlockUnderAnvil 函数来恢复承重方块所处位置的原本方块
*/
func (g *GlobalAPI) GenerateNewAnvil(pos [3]int32, blockStates string) (uuid.UUID, [3]int32, error) {
	resp, err := g.SendWSCommandWithResponce("querytarget @s")
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	got, err := g.ParseQuerytargetInfo(resp)
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
	uniqueId, err := uuid.NewUUID()
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	resp, err = g.SendWSCommandWithResponce(fmt.Sprintf(`structure save "%v" %d %d %d %d %d %d`, uniqueId.String(), pos[0], pos[1]-1, pos[2], pos[0], pos[1], pos[2]))
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	if resp.SuccessCount <= 0 {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: Failed to save blocks under %v; resp = %#v", pos, resp)
	}
	// 备份相关的方块
	err = g.SendSettingsCommand(fmt.Sprintf("setblock %d %d %d %v", pos[0], pos[1]-1, pos[2], BlockUnderAnvil), true)
	if err != nil {
		return uuid.UUID{}, [3]int32{}, fmt.Errorf("GenerateNewAnvil: %v", err)
	}
	resp, err = g.SendWSCommandWithResponce(fmt.Sprintf("setblock %d %d %d anvil %v", pos[0], pos[1], pos[2], blockStates))
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

// 恢复铁砧及对应承重方块处的方块为原本方块，同时删除备份用结构。
// 其中，uniqueId 参数代表备份用结构的名称在被转换为 uuid.UUID 后的结果。
// 特别地，anvilPos 参数应当填写铁砧被放置的坐标。
func (g *GlobalAPI) RevertBlocks(uniqueId uuid.UUID, anvilPos [3]int32) error {
	correctPos := [3]int32{anvilPos[0], anvilPos[1] - 1, anvilPos[2]}
	// 初始化
	resp, err := g.SendWSCommandWithResponce(fmt.Sprintf(`structure load "%v" %d %d %d`, uniqueId.String(), correctPos[0], correctPos[1], correctPos[2]))
	if err != nil {
		return fmt.Errorf("RevertBlocks: %v", err)
	}
	if resp.SuccessCount <= 0 {
		return fmt.Errorf("RevertBlocks: Failed to revert blocks on %v; resp = %#v", correctPos, resp)
	}
	// 尝试恢复承重方块处原本的方块
	err = g.SendSettingsCommand(fmt.Sprintf(`structure delete "%v"`, uniqueId.String()), true)
	if err != nil {
		return fmt.Errorf("RevertBlocks: %v", err)
	}
	// 删除用于备份的结构
	return nil
	// 返回值
}
