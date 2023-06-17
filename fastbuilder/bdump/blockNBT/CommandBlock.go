package blockNBT

import (
	"fmt"
	"phoenixbuilder/fastbuilder/commands_generator"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/minecraft/protocol/packet"
)

// CommandBlockDatas 结构体用于描述命令方块的 NBT 在被解析后的数据
type CommandBlockDatas struct {
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
	Package *Package
	// 命令方块数据
	CommandBlockDatas CommandBlockDatas
	// 为向下兼容而设，因为旧方法下不需要放置命令方块
	NeedToPlaceBlock bool
}

// 从 c.Package.Block.NBT 提取命令方块数据并保存在 c.CommandBlockDatas 中
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
	_, ok := c.Package.Block.NBT["Command"]
	if ok {
		command, normal = c.Package.Block.NBT["Command"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.Package.Block.NBT[\"Command\"]; c.Package.Block.NBT = %#v", c.Package.Block.NBT)
		}
	}
	// Command
	_, ok = c.Package.Block.NBT["CustomName"]
	if ok {
		customName, normal = c.Package.Block.NBT["CustomName"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.Package.Block.NBT[\"CustomName\"]; c.Package.Block.NBT = %#v", c.Package.Block.NBT)
		}
	}
	// CustomName
	_, ok = c.Package.Block.NBT["LastOutput"]
	if ok {
		lastOutput, normal = c.Package.Block.NBT["LastOutput"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.Package.Block.NBT[\"LastOutput\"]; c.Package.Block.NBT = %#v", c.Package.Block.NBT)
		}
	}
	// LastOutput
	_, ok = c.Package.Block.NBT["TickDelay"]
	if ok {
		tickDelay, normal = c.Package.Block.NBT["TickDelay"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.Package.Block.NBT[\"TickDelay\"]; c.Package.Block.NBT = %#v", c.Package.Block.NBT)
		}
	}
	// TickDelay
	_, ok = c.Package.Block.NBT["ExecuteOnFirstTick"]
	if ok {
		got, normal := c.Package.Block.NBT["ExecuteOnFirstTick"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.Package.Block.NBT[\"ExecuteOnFirstTick\"]; c.Package.Block.NBT = %#v", c.Package.Block.NBT)
		}
		if got == byte(0) {
			executeOnFirstTick = false
		} else {
			executeOnFirstTick = true
		}
	}
	// ExecuteOnFirstTick
	_, ok = c.Package.Block.NBT["TrackOutput"]
	if ok {
		got, normal := c.Package.Block.NBT["TrackOutput"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.Package.Block.NBT[\"TrackOutput\"]; c.Package.Block.NBT = %#v", c.Package.Block.NBT)
		}
		if got == byte(0) {
			trackOutput = false
		} else {
			trackOutput = true
		}
	}
	// TrackOutput
	_, ok = c.Package.Block.NBT["conditionalMode"]
	if ok {
		got, normal := c.Package.Block.NBT["conditionalMode"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.Package.Block.NBT[\"conditionalMode\"]; c.Package.Block.NBT = %#v", c.Package.Block.NBT)
		}
		if got == byte(0) {
			conditionalMode = false
		} else {
			conditionalMode = true
		}
	}
	// conditionalMode
	_, ok = c.Package.Block.NBT["auto"]
	if ok {
		got, normal := c.Package.Block.NBT["auto"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.Package.Block.NBT[\"auto\"]; c.Package.Block.NBT = %#v", c.Package.Block.NBT)
		}
		if got == byte(0) {
			auto = false
		} else {
			auto = true
		}
	}
	// auto
	c.CommandBlockDatas = CommandBlockDatas{
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
func (c *CommandBlock) WriteDatas() error {
	var mode uint32 = packet.CommandBlockImpulse
	// 初始化
	if c.NeedToPlaceBlock {
		if c.Package.Datas.Settings.ExcludeCommands || c.Package.Datas.FastMode {
			err := c.Package.API.SetBlockFastly(c.Package.Datas.Position, c.Package.Block.Name, c.Package.Datas.StatesString)
			if err != nil {
				return fmt.Errorf("WriteDatas: %v", err)
			}
			// 如果要求仅放置命令方块亦或以快速模式放置命令方块
		} else {
			err := c.Package.API.SetBlock(c.Package.Datas.Position, c.Package.Block.Name, c.Package.Datas.StatesString)
			if err != nil {
				return fmt.Errorf("WriteDatas: %v", err)
			}
			// 普通情形
		}
	}
	// 放置命令方块
	if c.Package.Datas.Settings.ExcludeCommands {
		return nil
	}
	// 如果不要求写入命令方块数据
	err := c.Package.API.SendSettingsCommand(fmt.Sprintf("tp %d %d %d", c.Package.Datas.Position[0], c.Package.Datas.Position[1], c.Package.Datas.Position[2]), true)
	if err != nil {
		return fmt.Errorf("WriteDatas: %v", err)
	}
	// 传送机器人到命令方块位置
	if c.Package.Block.Name == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else if c.Package.Block.Name == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	}
	// 确定命令方块的类型
	if c.Package.Datas.Settings.InvalidateCommands {
		c.CommandBlockDatas.Command = "# " + c.CommandBlockDatas.Command
	}
	// 如果需要对命令无效化处理
	err = c.Package.API.WritePacket(&packet.CommandBlockUpdate{
		Block:              true,
		Position:           c.Package.Datas.Position,
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

// 以旧方法放置命令方块；主要用于向下兼容，如 operation 36 等
func (c *CommandBlock) PlaceCommandBlockWithLegacyMethod(
	block *types.Module,
	cfg *types.MainConfig,
) error {
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
	if block.Block == nil {
		block.Block = &types.Block{}
		block.Block.Name = &blockName

		{
			_, err := c.Package.API.SendWSCommandWithResponce("list")
			if err != nil {
				return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
			}
		}
		// 这么做的目的只是为了保证存在 operation 26 - SetCommandBlockData 的时候，
		// 命令方块导入速度不会太快。
		// 不过这个解决方案不是很优雅，
		// 但这里并没有 GetBlockUpdateSubscribeMap 类似的实现，
		// 所以暂且先这样吧

		err := c.WriteDatas()
		if err != nil {
			return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
		}
		return nil
	}
	block.Block.Name = &blockName
	c.Package.Block.Name = blockName
	// 确定命令方块的类型 & 如果是 operation 26 - SetCommandBlockData
	request := commands_generator.SetBlockRequest(block, cfg)
	if c.Package.Datas.FastMode {
		err := c.Package.API.SendSettingsCommand(request, true)
		if err != nil {
			return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
		}
	} else {
		_, err := c.Package.API.SendWSCommandWithResponce(request)
		if err != nil {
			return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
		}
	}
	// 放置命令方块
	err := c.WriteDatas()
	if err != nil {
		return fmt.Errorf("PlaceCommandBlockWithLegacyMethod: %v", err)
	}
	// 写入命令方块数据
	return nil
	// 返回值
}
