package blockNBT_CommandBlock

import (
	"fmt"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 放置一个命令方块(可选)并写入命令方块数据
func (c *CommandBlock) WriteDatas(needToPlaceBlock bool) error {
	var mode uint32 = packet.CommandBlockImpulse
	// 初始化
	if needToPlaceBlock {
		if c.BlockEntityDatas.Datas.Settings.ExcludeCommands || c.BlockEntityDatas.Datas.FastMode {
			err := c.BlockEntityDatas.API.SetBlockFastly(c.BlockEntityDatas.Datas.Position, c.BlockEntityDatas.Block.Name, c.BlockEntityDatas.Datas.StatesString)
			if err != nil {
				return fmt.Errorf("WriteDatas: %v", err)
			}
			// 如果要求仅放置命令方块亦或以快速模式放置命令方块
		} else {
			err := c.BlockEntityDatas.API.SetBlock(c.BlockEntityDatas.Datas.Position, c.BlockEntityDatas.Block.Name, c.BlockEntityDatas.Datas.StatesString)
			if err != nil {
				return fmt.Errorf("WriteDatas: %v", err)
			}
			// 普通情形
		}
	}
	// 放置命令方块
	if c.BlockEntityDatas.Datas.Settings.ExcludeCommands {
		return nil
	}
	// 如果不要求写入命令方块数据
	err := c.BlockEntityDatas.API.SendSettingsCommand(fmt.Sprintf("tp %d %d %d", c.BlockEntityDatas.Datas.Position[0], c.BlockEntityDatas.Datas.Position[1], c.BlockEntityDatas.Datas.Position[2]), true)
	if err != nil {
		return fmt.Errorf("WriteDatas: %v", err)
	}
	// 传送机器人到命令方块位置
	if c.BlockEntityDatas.Block.Name == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else if c.BlockEntityDatas.Block.Name == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	}
	// 确定命令方块的类型
	if c.BlockEntityDatas.Datas.Settings.InvalidateCommands {
		c.CommandBlockDatas.Command = "# " + c.CommandBlockDatas.Command
	}
	// 如果需要对命令无效化处理
	err = c.BlockEntityDatas.API.WritePacket(&packet.CommandBlockUpdate{
		Block:              true,
		Position:           c.BlockEntityDatas.Datas.Position,
		Mode:               mode,
		NeedsRedstone:      !c.CommandBlockDatas.Auto,
		Conditional:        c.CommandBlockDatas.ConditionalMode,
		Command:            c.CommandBlockDatas.Command,
		LastOutput:         c.CommandBlockDatas.LastOutput,
		Name:               c.CommandBlockDatas.CustomName,
		ShouldTrackOutput:  c.CommandBlockDatas.TrackOutput,
		TickDelay:          c.CommandBlockDatas.TickDelay,
		ExecuteOnFirstTick: c.CommandBlockDatas.ExecuteOnFirstTick,
	})
	if err != nil {
		return fmt.Errorf("WriteDatas: %v", err)
	}
	// 写入命令方块数据
	return nil
	// 返回值
}
