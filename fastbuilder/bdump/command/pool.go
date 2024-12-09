package command

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

var BDumpCommandPool map[uint16]func() Command = map[uint16]func() Command{
	1:  func() Command { return &CreateConstantString{} },
	5:  func() Command { return &PlaceBlockWithBlockStates{} },
	6:  func() Command { return &AddInt16ZValue0{} },
	7:  func() Command { return &PlaceBlock{} },
	8:  func() Command { return &AddZValue0{} },
	9:  func() Command { return &NoOperation{} },
	12: func() Command { return &AddInt32ZValue0{} },
	13: func() Command { return &PlaceBlockWithBlockStatesDeprecated{} },
	14: func() Command { return &AddXValue{} },
	15: func() Command { return &SubtractXValue{} },
	16: func() Command { return &AddYValue{} },
	17: func() Command { return &SubtractYValue{} },
	18: func() Command { return &AddZValue{} },
	19: func() Command { return &SubtractZValue{} },
	20: func() Command { return &AddInt16XValue{} },
	21: func() Command { return &AddInt32XValue{} },
	22: func() Command { return &AddInt16YValue{} },
	23: func() Command { return &AddInt32YValue{} },
	24: func() Command { return &AddInt16ZValue{} },
	25: func() Command { return &AddInt32ZValue{} },
	26: func() Command { return &SetCommandBlockData{} },
	27: func() Command { return &PlaceBlockWithCommandBlockData{} },
	28: func() Command { return &AddInt8XValue{} },
	29: func() Command { return &AddInt8YValue{} },
	30: func() Command { return &AddInt8ZValue{} },
	31: func() Command { return &UseRuntimeIDPool{} },
	32: func() Command { return &PlaceRuntimeBlock{} },
	33: func() Command { return &PlaceRuntimeBlockWithUint32RuntimeID{} },
	34: func() Command { return &PlaceRuntimeBlockWithCommandBlockData{} },
	35: func() Command { return &PlaceRuntimeBlockWithCommandBlockDataAndUint32RuntimeID{} },
	36: func() Command { return &PlaceCommandBlockWithCommandBlockData{} },
	37: func() Command { return &PlaceRuntimeBlockWithChestData{} },
	38: func() Command { return &PlaceRuntimeBlockWithChestDataAndUint32RuntimeID{} },
	39: func() Command { return &AssignDebugData{} },
	40: func() Command { return &PlaceBlockWithChestData{} },
	41: func() Command { return &PlaceBlockWithNBTData{} },
	88: func() Command { return &Terminate{} },
}
