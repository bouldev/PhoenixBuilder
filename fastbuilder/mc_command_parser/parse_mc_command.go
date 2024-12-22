package mc_command_parser

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
	"strconv"
)

// 从 command 解析一个 execute 命令。
// 若返回 nil ，则 command 不是一个 execute 命令
func ParseExecuteCommand(command string) (e *ExecuteCommand) {
	p := NewCommandParser(&command)
	r := p.reader
	// prepare
	if p.ExpectHeader("execute", true) {
		e = &ExecuteCommand{}
	} else {
		return
	}
	// check header
	r.JumpSpace()
	e.Selector = p.ParseSelector()
	// parse selector
	r.JumpSpace()
	e.Position = p.ParsePosition()
	// parse block position
	r.JumpSpace()
	if p.ExpectHeader("detect", false) {
		r.JumpSpace()
		tmp := p.ParseDetectArgs()
		e.DetectArgs = &tmp
	}
	// parse detect args
	r.JumpSpace()
	e.SubCommand = command[r.Pointer():]
	// get sub command
	return
	// return
}

// 从 blockStates 解析一组方块状态。
// 若返回 nil ，则 blockStates 不是一个方块状态
func ParseBlockStates(blockStates string) (m map[string]interface{}) {
	version := 0
	p := NewCommandParser(&blockStates)
	r := p.reader
	// prepare
	r.JumpSpace()
	if r.Next(true) == "[" {
		m = make(map[string]interface{})
	} else {
		return
	}
	// check header
	r.JumpSpace()
	switch r.Next(false) {
	case "]":
		return
	default:
		r.SetPtr(r.Pointer() - 1)
	}
	// when the block state is empty
	for {
		r.JumpSpace()
		if r.Next(false) != `"` {
			panic("ParseBlockStates: Invalid block states string")
		}
		key := r.ParseString()
		// get key
		r.JumpSpace()
		switch r.Next(false) {
		case ":":
			if version == 2 {
				panic("ParseBlockStates: Invalid block states string")
			}
			version = 1
		case "=":
			if version == 1 {
				panic("ParseBlockStates: Invalid block states string")
			}
			version = 2
		default:
			panic("ParseBlockStates: Invalid block states string")
		}
		// check separator
		r.JumpSpace()
		switch r.Next(false) {
		case "+", "-", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			r.SetPtr(r.Pointer() - 1)
			intString, isInt := r.ParseNumber(true)
			if !isInt {
				panic("ParseBlockStates: The value of the key provided can not be a float")
			}
			num, err := strconv.ParseInt(intString, 10, 32)
			if err != nil {
				panic(fmt.Sprintf("ParseBlockStates: %v", err))
			}
			m[key] = int32(num)
			// e.g. `+23`
		case `"`:
			m[key] = r.ParseString()
			// e.g. `"\\\"233\"\\"`
		case "t", "f", "T", "F":
			r.SetPtr(r.Pointer() - 1)
			boolean := r.ParseBool()
			if boolean {
				m[key] = byte(1)
			} else {
				m[key] = byte(0)
			}
			// e.g. `TrUE`, "FALSE"
		default:
			panic("ParseBlockStates: Invalid block states string")
			// unknown format
		}
		// get value of the key
		r.JumpSpace()
		switch r.Next(false) {
		case ",":
		case "]":
			return
		default:
			panic("ParseBlockStates: Invalid block states string")
		}
		// prepare for the next state
	}
	// parse each state and return
}
