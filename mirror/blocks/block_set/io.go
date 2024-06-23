package block_set

import (
	"bufio"
	"bytes"
	"fmt"
	"phoenixbuilder/mirror/blocks/describe"
	"strconv"
	"strings"
)

func BlockSetFromStringRecords(dataBytes string, unknownRtid uint32) *BlockSet {
	blocks := &BlockSet{
		unknownRuntimeID: unknownRtid,
	}
	reader := bufio.NewReader(bytes.NewBufferString(dataBytes))
	{
		version, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if !strings.HasPrefix(version, "VERSION:") {
			panic(fmt.Errorf("expect VERSION:, get %v", version))
		} else {
			versionInt, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(version, "VERSION:")))
			blocks.version = uint32(versionInt)
			if blocks.version == 0 {
				panic(fmt.Errorf("cannot get nemc block version: %v", version))
			}
		}
	}
	numBlocks := 0
	{
		count, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if !strings.HasPrefix(count, "COUNTS:") {
			panic(fmt.Errorf("expect COUNTS:, get %v", count))
		} else {
			numBlocks, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(count, "COUNTS:")))
			if numBlocks == 0 {
				panic(fmt.Errorf("cannot get block count: %v", count))
			} else {
				// fmt.Printf("nemc block count: %v\n", numBlocks)
			}
		}
	}
	bs := make([]*describe.Block, 0, numBlocks)
	for runtimeID := int32(0); runtimeID < int32(numBlocks); runtimeID++ {
		blockName, err := reader.ReadString(' ')
		if err != nil {
			panic(err)
		}
		blockName = strings.TrimSpace(blockName)
		blockValStr, err := reader.ReadString(' ')
		if err != nil {
			panic(err)
		}
		blockValStr = strings.TrimSpace(blockValStr)
		blockVal, err := strconv.Atoi(blockValStr)
		if err != nil {
			panic(err)
		}
		statesSnbt, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		statesSnbt = strings.TrimSpace(statesSnbt)
		b := describe.NewBlockFromSnbt(blockName, statesSnbt, uint16(blockVal), uint32(runtimeID))

		if b.States().SNBTString() != statesSnbt {
			panic(fmt.Errorf("snbt error: %v!=%v", b.States().SNBTString(), statesSnbt))
		}
		if blockName == "air" {
			blocks.airRuntimeID = uint32(runtimeID)
		}
		bs = append(bs, b)
	}
	if blocks.airRuntimeID == 0 {
		panic("cannot found air runtime id")
	}
	blocks.blocks = bs
	if blocks.blocks[blocks.airRuntimeID].ShortName() != "air" {
		panic("should not happen")
	}
	if blocks.version == 0 {
		panic("no version info")
	}
	if blocks.unknownRuntimeID == 0 {
		panic("unknown runtime id not set")
	}
	// fmt.Println(len(blocks.blocks))
	return blocks
}

func (blocks *BlockSet) DumpStringRecords() string {
	datas := fmt.Sprintf("VERSION:%v\nCOUNTS:%v\n", blocks.version, len(blocks.blocks))
	for _, block := range blocks.blocks {
		datas += fmt.Sprintf("%v %v %v\n", block.ShortName(), block.LegacyValue(), block.States().SNBTString())
	}
	return datas
}
