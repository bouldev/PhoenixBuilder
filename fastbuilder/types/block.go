package types

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

type Module struct {
	Block            *Block
	CommandBlockData *CommandBlockData
	DebugNBTData     []byte
	NBTData          []byte
	NBTMap           map[string]interface{}
	ChestSlot        *ChestSlot
	ChestData        *ChestData
	Point            Position
}

type RuntimeModule struct {
	BlockRuntimeId   uint32 // The current total count of runtime ids didn't exceed 65536
	CommandBlockData *CommandBlockData
	ChestData        *ChestData
	Point            Position
}

type Block struct {
	Name        *string
	BlockStates string
	Data        uint16
}

type CommandBlockData struct {
	Mode               uint32
	Command            string
	CustomName         string
	LastOutput         string
	TickDelay          int32
	ExecuteOnFirstTick bool //byte
	TrackOutput        bool //byte
	Conditional        bool
	NeedsRedstone      bool
}

type ChestData []ChestSlot

type ChestSlot struct {
	Name   string
	Count  uint8
	Damage uint16
	Slot   uint8
}

type ConstBlock struct {
	Name string
	Data uint16
}

type DoubleModule struct {
	Begin           Position
	End             Position
	Block, OldBlock *Block
}

var takenBlocks map[*ConstBlock]*Block = make(map[*ConstBlock]*Block)

const takenBlocksMaxSize = 1024
const takenBlocksDeleteCount = 512

func CreateBlock(name string, data uint16) *Block {
	return &Block{
		Name: &name,
		Data: data,
	}
}

func (req *ConstBlock) Take() *Block {
	block, ok := takenBlocks[req]
	if ok {
		return block
	}
	if len(takenBlocks) > takenBlocksMaxSize {
		i := 0
		for k, _ := range takenBlocks {
			delete(takenBlocks, k)
			i++
			if i >= takenBlocksDeleteCount {
				break
			}
		}
	}
	block = &Block{
		Name: &req.Name, //ConstBlock won't be destroyed
		Data: req.Data,
	}
	takenBlocks[req] = block
	return block
}
