package blockNBT_CommandBlock

import "fmt"

// 从 c.BlockEntityDatas.Block.NBT 提取命令方块数据并保存在 c.CommandBlockDatas 中
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
	_, ok := c.BlockEntityDatas.Block.NBT["Command"]
	if ok {
		command, normal = c.BlockEntityDatas.Block.NBT["Command"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.BlockEntityDatas.Block.NBT[\"Command\"]; c.BlockEntityDatas.Block.NBT = %#v", c.BlockEntityDatas.Block.NBT)
		}
	}
	// Command
	_, ok = c.BlockEntityDatas.Block.NBT["CustomName"]
	if ok {
		customName, normal = c.BlockEntityDatas.Block.NBT["CustomName"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.BlockEntityDatas.Block.NBT[\"CustomName\"]; c.BlockEntityDatas.Block.NBT = %#v", c.BlockEntityDatas.Block.NBT)
		}
	}
	// CustomName
	_, ok = c.BlockEntityDatas.Block.NBT["LastOutput"]
	if ok {
		lastOutput, normal = c.BlockEntityDatas.Block.NBT["LastOutput"].(string)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.BlockEntityDatas.Block.NBT[\"LastOutput\"]; c.BlockEntityDatas.Block.NBT = %#v", c.BlockEntityDatas.Block.NBT)
		}
	}
	// LastOutput
	_, ok = c.BlockEntityDatas.Block.NBT["TickDelay"]
	if ok {
		tickDelay, normal = c.BlockEntityDatas.Block.NBT["TickDelay"].(int32)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.BlockEntityDatas.Block.NBT[\"TickDelay\"]; c.BlockEntityDatas.Block.NBT = %#v", c.BlockEntityDatas.Block.NBT)
		}
	}
	// TickDelay
	_, ok = c.BlockEntityDatas.Block.NBT["ExecuteOnFirstTick"]
	if ok {
		got, normal := c.BlockEntityDatas.Block.NBT["ExecuteOnFirstTick"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.BlockEntityDatas.Block.NBT[\"ExecuteOnFirstTick\"]; c.BlockEntityDatas.Block.NBT = %#v", c.BlockEntityDatas.Block.NBT)
		}
		if got == byte(0) {
			executeOnFirstTick = false
		} else {
			executeOnFirstTick = true
		}
	}
	// ExecuteOnFirstTick
	_, ok = c.BlockEntityDatas.Block.NBT["TrackOutput"]
	if ok {
		got, normal := c.BlockEntityDatas.Block.NBT["TrackOutput"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.BlockEntityDatas.Block.NBT[\"TrackOutput\"]; c.BlockEntityDatas.Block.NBT = %#v", c.BlockEntityDatas.Block.NBT)
		}
		if got == byte(0) {
			trackOutput = false
		} else {
			trackOutput = true
		}
	}
	// TrackOutput
	_, ok = c.BlockEntityDatas.Block.NBT["conditionalMode"]
	if ok {
		got, normal := c.BlockEntityDatas.Block.NBT["conditionalMode"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.BlockEntityDatas.Block.NBT[\"conditionalMode\"]; c.BlockEntityDatas.Block.NBT = %#v", c.BlockEntityDatas.Block.NBT)
		}
		if got == byte(0) {
			conditionalMode = false
		} else {
			conditionalMode = true
		}
	}
	// conditionalMode
	_, ok = c.BlockEntityDatas.Block.NBT["auto"]
	if ok {
		got, normal := c.BlockEntityDatas.Block.NBT["auto"].(byte)
		if !normal {
			return fmt.Errorf("Decode: Crashed in c.BlockEntityDatas.Block.NBT[\"auto\"]; c.BlockEntityDatas.Block.NBT = %#v", c.BlockEntityDatas.Block.NBT)
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
