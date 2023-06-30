package blockNBT

import (
	"fmt"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/types"
	GameInterface "phoenixbuilder/game_control/game_interface"
	"phoenixbuilder/minecraft/protocol/packet"
)

// Parsed Command block NBT data
type CommandBlockData struct {
	Command            string // Command(TAG_String) = ""
	CustomName         string // CustomName(TAG_String) = ""
	LastOutput         string // LastOutput(TAG_String) = ""
	TickDelay          int32  // TickDelay(TAG_Int) = 0
	ExecuteOnFirstTick bool   // ExecuteOnFirstTick(TAG_Byte) = 1
	TrackOutput        bool   // TrackOutput(TAG_Byte) = 1
	ConditionalMode    bool   // conditionalMode(TAG_Byte) = 0
	Auto               bool   // auto(TAG_Byte) = 1
}

// CommandBlock 结构体用于描述一个完整的命令方块数据
type CommandBlock struct {
	// 该方块实体的详细数据
	BlockEntity *BlockEntity
	CommandBlockData
	// 为向下兼容而设，因为旧方法下不需要放置命令方块
	ShouldPlaceBlock bool
}

// 从 c.Package.Block.NBT 提取命令方块数据并保存在 c.CommandBlockData 中
func (c *CommandBlock) Decode() error {
	var normal bool = false
	var command string = ""
	var customName string = ""
	var lastOutput string = ""
	var tickDelay int32 = int32(0)
	var executeOnFirstTick bool = true
	var trackOutput bool = true
	var conditionalMode bool = false
	var auto bool = true
	// 初始化
	_, ok := c.BlockEntity.Block.NBT["Command"]
	if ok {
		command, normal = c.BlockEntity.Block.NBT["Command"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed at c.BlockEntity.Block.NBT[\"Command\"]; c.BlockEntity.Block.NBT = %#v", c.BlockEntity.Block.NBT)
		}
	}
	// Command
	_, ok = c.BlockEntity.Block.NBT["CustomName"]
	if ok {
		customName, normal = c.BlockEntity.Block.NBT["CustomName"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed at c.BlockEntity.Block.NBT[\"CustomName\"]; c.Package.Block.NBT = %#v", c.BlockEntity.Block.NBT)
		}
	}
	// CustomName
	_, ok = c.BlockEntity.Block.NBT["LastOutput"]
	if ok {
		lastOutput, normal = c.BlockEntity.Block.NBT["LastOutput"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed at c.BlockEntity.Block.NBT[\"LastOutput\"]; c.BlockEntity.Block.NBT = %#v", c.BlockEntity.Block.NBT)
		}
	}
	// LastOutput
	_, ok = c.BlockEntity.Block.NBT["TickDelay"]
	if ok {
		tickDelay, normal = c.BlockEntity.Block.NBT["TickDelay"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.Package.Block.NBT[\"TickDelay\"]; c.BlockEntity.Block.NBT = %#v", c.BlockEntity.Block.NBT)
		}
	}
	// TickDelay
	_, ok = c.BlockEntity.Block.NBT["ExecuteOnFirstTick"]
	if ok {
		got, normal := c.BlockEntity.Block.NBT["ExecuteOnFirstTick"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed at c.BlockEntity.Block.NBT[\"ExecuteOnFirstTick\"]; c.BlockEntity.Block.NBT = %#v", c.BlockEntity.Block.NBT)
		}
		if got == byte(0) {
			executeOnFirstTick = false
		} else {
			executeOnFirstTick = true
		}
	}
	// ExecuteOnFirstTick
	_, ok = c.BlockEntity.Block.NBT["TrackOutput"]
	if ok {
		got, normal := c.BlockEntity.Block.NBT["TrackOutput"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed at c.BlockEntity.Block.NBT[\"TrackOutput\"]; c.BlockEntity.Block.NBT = %#v", c.BlockEntity.Block.NBT)
		}
		if got == byte(0) {
			trackOutput = false
		} else {
			trackOutput = true
		}
	}
	// TrackOutput
	_, ok = c.BlockEntity.Block.NBT["conditionalMode"]
	if ok {
		got, normal := c.BlockEntity.Block.NBT["conditionalMode"].(byte)
		if !normal {
			return fmt.Errorf("Decode: ERR 106; NBT = %#v", c.BlockEntity.Block.NBT)
		}
		if got == byte(0) {
			conditionalMode = false
		} else {
			conditionalMode = true
		}
	}
	// conditionalMode
	_, ok = c.BlockEntity.Block.NBT["auto"]
	if ok {
		got, normal := c.BlockEntity.Block.NBT["auto"].(byte)
		if !normal {
			return fmt.Errorf("Decode: ERR 333; NBT = %#v", c.BlockEntity.Block.NBT)
		}
		if got == byte(0) {
			auto = false
		} else {
			auto = true
		}
	}
	// auto
	c.CommandBlockData = CommandBlockData{
		Command:            command,
		CustomName:         customName,
		LastOutput:         lastOutput,
		TickDelay:          tickDelay,
		ExecuteOnFirstTick: executeOnFirstTick,
		TrackOutput:        trackOutput,
		ConditionalMode:    conditionalMode,
		Auto:               auto,
	}
	return nil
	// return
}

// 放置一个命令方块(可选)并写入命令方块数据
func (c *CommandBlock) WriteData() error {
	var mode uint32 = packet.CommandBlockImpulse
	gameInterface := c.BlockEntity.Interface.(*GameInterface.GameInterface)
	if c.ShouldPlaceBlock {
		if c.BlockEntity.AdditionalData.Settings.ExcludeCommands || c.BlockEntity.AdditionalData.FastMode {
			err := c.BlockEntity.Interface.SetBlockAsync(c.BlockEntity.AdditionalData.Position, c.BlockEntity.Block.Name, c.BlockEntity.AdditionalData.BlockStates)
			if err != nil {
				return fmt.Errorf("WriteData: %v", err)
			}
			// 如果要求仅放置命令方块亦或以快速模式放置命令方块
		} else {
			err := c.BlockEntity.Interface.SetBlock(c.BlockEntity.AdditionalData.Position, c.BlockEntity.Block.Name, c.BlockEntity.AdditionalData.BlockStates)
			if err != nil {
				return fmt.Errorf("WriteData: %v", err)
			}
			// 普通情形
		}
	}
	// 放置命令方块
	if c.BlockEntity.AdditionalData.Settings.ExcludeCommands {
		return nil
	}
	// 如果不要求写入命令方块数据
	err := c.BlockEntity.Interface.SendSettingsCommand(fmt.Sprintf("tp %d %d %d", c.BlockEntity.AdditionalData.Position[0], c.BlockEntity.AdditionalData.Position[1], c.BlockEntity.AdditionalData.Position[2]), true)
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	if c.BlockEntity.Block.Name == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else if c.BlockEntity.Block.Name == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	}
	if c.BlockEntity.AdditionalData.Settings.InvalidateCommands {
		c.CommandBlockData.Command = "# " + c.CommandBlockData.Command
	}
	err = gameInterface.WritePacket(&packet.CommandBlockUpdate{
		Block:              true,
		Position:           c.BlockEntity.AdditionalData.Position,
		Mode:               mode,
		NeedsRedstone:      !c.CommandBlockData.Auto,
		Conditional:        c.CommandBlockData.ConditionalMode,
		Command:            c.CommandBlockData.Command,
		LastOutput:         c.CommandBlockData.LastOutput,
		Name:               c.CommandBlockData.CustomName,
		ShouldTrackOutput:  c.CommandBlockData.TrackOutput,
		TickDelay:          c.CommandBlockData.TickDelay,
		ExecuteOnFirstTick: c.CommandBlockData.ExecuteOnFirstTick,
	})
	if err != nil {
		return fmt.Errorf("WriteData: %v", err)
	}
	// 写入命令方块数据
	return nil
	// 返回值
}

// 以旧方法放置命令方块；主要用于向下兼容，如 operation 36 等
func (c *CommandBlock) PlaceCommandBlockLegacy(
	block *types.Module,
	cfg *types.MainConfig,
) error {
	var blockName string = "command_block"
	c.CommandBlockData = CommandBlockData{
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
	if block.Block == nil {
		block.Block = &types.Block{}
		block.Block.Name = &blockName

		// TODO: 优化下方的这一段代码
		{
			_, err := c.BlockEntity.Interface.(*GameInterface.GameInterface).SendWSCommandWithResponse("list")
			if err != nil {
				return fmt.Errorf("PlaceCommandBlockLegacy: %v", err)
			}
		}
		// 这么做的目的只是为了保证存在 operation 26 - SetCommandBlockData 的时候，
		// 命令方块导入速度不会太快。
		// 不过这个解决方案不是很优雅，
		// 但这里并没有 GetBlockUpdateSubscribeMap 类似的实现，
		// 所以暂且先这样吧

		err := c.WriteData()
		if err != nil {
			return fmt.Errorf("PlaceCommandBlockLegacy: %v", err)
		}
		return nil
	}
	block.Block.Name = &blockName
	c.BlockEntity.Block.Name = blockName
	// 确定命令方块的类型 & 如果是 operation 26 - SetCommandBlockData
	request := commands_generator.SetBlockRequest(block, cfg)
	if c.BlockEntity.AdditionalData.FastMode {
		err := c.BlockEntity.Interface.SendSettingsCommand(request, true)
		if err != nil {
			return fmt.Errorf("ERR 444eee %v", err)
		}
	} else {
		_, err := c.BlockEntity.Interface.(*GameInterface.GameInterface).SendWSCommandWithResponse(request)
		if err != nil {
			return fmt.Errorf("ERR 555ccc %v", err)
		}
	}
	// 放置命令方块
	err := c.WriteData()
	if err != nil {
		return fmt.Errorf("PlaceCommandBlockLegacy: %v", err)
	}
	// 写入命令方块数据
	return nil
	// 返回值
}
