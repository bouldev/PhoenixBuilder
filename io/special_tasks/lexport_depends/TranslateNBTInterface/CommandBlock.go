package TranslateNBTInerface

import (
	"fmt"
	"phoenixbuilder/fastbuilder/types"
)

func GetCommandBlockData(cb map[string]interface{}, blockName string) (types.CommandBlockData, error) {
	var normal bool = false
	var command string = ""
	var customName string = ""
	var lastOutput string = ""
	var mode int = 0
	var tickDelay int32 = int32(0)
	var executeOnFirstTick bool = true
	var trackOutput bool = true
	var conditionalMode bool = false
	var needRedstone bool = true
	// 初始化
	_, ok := cb["Command"]
	if ok {
		command, normal = cb["Command"].(string)
		if !normal {
			return types.CommandBlockData{}, fmt.Errorf("Crashed in input[\"Command\"]")
		}
	}
	// Command
	_, ok = cb["CustomName"]
	if ok {
		customName, normal = cb["CustomName"].(string)
		if !normal {
			return types.CommandBlockData{}, fmt.Errorf("Crashed in input[\"CustomName\"]")
		}
	}
	// CustomName
	_, ok = cb["LastOutput"]
	if ok {
		lastOutput, normal = cb["LastOutput"].(string)
		if !normal {
			return types.CommandBlockData{}, fmt.Errorf("Crashed in input[\"LastOutput\"]")
		}
	}
	// LastOutput
	if blockName == "command_block" {
		mode = 0
	} else if blockName == "repeating_command_block" {
		mode = 1
	} else if blockName == "chain_command_block" {
		mode = 2
	} else {
		return types.CommandBlockData{}, fmt.Errorf("Not a command block")
	}
	// mode
	_, ok = cb["TickDelay"]
	if ok {
		tickDelay, normal = cb["TickDelay"].(int32)
		if !normal {
			return types.CommandBlockData{}, fmt.Errorf("Crashed in input[\"TickDelay\"]")
		}
	}
	// TickDelay
	_, ok = cb["ExecuteOnFirstTick"]
	if ok {
		got, normal := cb["ExecuteOnFirstTick"].(byte)
		if !normal {
			return types.CommandBlockData{}, fmt.Errorf("Crashed in input[\"ExecuteOnFirstTick\"]")
		}
		if got == byte(0) {
			executeOnFirstTick = false
		} else {
			executeOnFirstTick = true
		}
	}
	// ExecuteOnFirstTick
	_, ok = cb["TrackOutput"]
	if ok {
		got, normal := cb["TrackOutput"].(byte)
		if !normal {
			return types.CommandBlockData{}, fmt.Errorf("Crashed in input[\"TrackOutput\"]")
		}
		if got == byte(0) {
			trackOutput = false
		} else {
			trackOutput = true
		}
	}
	// TrackOutput
	_, ok = cb["conditionalMode"]
	if ok {
		got, normal := cb["conditionalMode"].(byte)
		if !normal {
			return types.CommandBlockData{}, fmt.Errorf("Crashed in input[\"conditionalMode\"]")
		}
		if got == byte(0) {
			conditionalMode = false
		} else {
			conditionalMode = true
		}
	}
	// Conditional
	_, ok = cb["auto"]
	if ok {
		got, normal := cb["auto"].(byte)
		if !normal {
			return types.CommandBlockData{}, fmt.Errorf("Crashed in input[\"auto\"]")
		}
		if got == byte(0) {
			needRedstone = true
		} else {
			needRedstone = false
		}
	}
	// auto
	return types.CommandBlockData{
		Mode:               uint32(mode),
		Command:            command,
		CustomName:         customName,
		LastOutput:         lastOutput,
		TickDelay:          tickDelay,
		ExecuteOnFirstTick: executeOnFirstTick,
		TrackOutput:        trackOutput,
		Conditional:        conditionalMode,
		NeedsRedstone:       needRedstone,
	}, nil
}
