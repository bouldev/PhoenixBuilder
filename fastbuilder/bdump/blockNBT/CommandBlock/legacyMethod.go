package blockNBT_CommandBlock

import (
	"fmt"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/minecraft/protocol/packet"
)

// 以旧方法放置命令方块；主要用于向下兼容，如 operation 36 等
func (c *CommandBlock) PlaceCommandBlockWithLegacyMethod(block *types.Module, cfg *types.MainConfig) error {
	var blockName string = "command_block"
	c.CommandBlockDatas = CommandBlockDatas{
		Command:            block.CommandBlockData.Command,
		CustomName:         block.CommandBlockData.CustomName,
		LastOutput:         block.CommandBlockData.LastOutput,
		TickDelay:          block.CommandBlockData.TickDelay,
		ExecuteOnFirstTick: block.CommandBlockData.ExecuteOnFirstTick,
		TrackOutput:        block.CommandBlockData.TrackOutput,
		ConditionalMode:    block.CommandBlockData.Conditional,
		Auto:               !block.CommandBlockData.NeedsRedstone,
	}
	// 初始化
	if block.CommandBlockData.Mode == packet.CommandBlockChain {
		blockName = "chain_command_block"
	} else if block.CommandBlockData.Mode == packet.CommandBlockRepeating {
		blockName = "repeating_command_block"
	}
	block.Block.Name = &blockName
	c.BlockEntityDatas.Block.Name = blockName
	// 确定命令方块的类型
	if block.Block == nil {
		err := c.WriteDatas(false)
		if err != nil {
			return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
		}
		return nil
	}
	// 如果是 operation 26 - SetCommandBlockData
	request := commands_generator.SetBlockRequest(block, cfg)
	if c.BlockEntityDatas.Datas.FastMode {
		err := c.BlockEntityDatas.API.SendSettingsCommand(request, true)
		if err != nil {
			return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
		}
	} else {
		_, err := c.BlockEntityDatas.API.SendWSCommandWithResponce(request)
		if err != nil {
			return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
		}
	}
	// 放置命令方块
	err := c.WriteDatas(false)
	if err != nil {
		return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
	}
	// 写入命令方块数据
	return nil
	// 返回值
}
