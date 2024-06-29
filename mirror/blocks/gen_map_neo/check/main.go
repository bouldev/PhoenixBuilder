package main

import (
	"fmt"
	"phoenixbuilder/mirror/blocks"
)

func main() {
	// rtid, _ := chunk.LegacyBlockToRuntimeID("cartography_table", 0)

	// fmt.Println(rtid)
	// fmt.Println(chunk.RuntimeIDToLegacyBlock(rtid))

	// rtid, _ = chunk.LegacyBlockToRuntimeID("coral_block", 4)
	// fmt.Println(rtid)
	// blockName, state, _ := chunk.RuntimeIDToState(rtid)
	// fmt.Println(blockName, state)
	// blockNameWithStateStr, _ := chunk.RuntimeIDToStateStr(rtid)
	// fmt.Println(blockNameWithStateStr)
	// blockName, StateStr, _ := chunk.RuntimeIDToNameAndStateStr(rtid)
	// fmt.Println(blockName, StateStr)
	// fmt.Println(chunk.StateToRuntimeID(blockName, state))
	// fmt.Println(chunk.BlockStateStrToRuntimeID(blockName, StateStr))
	// rtid, _ = chunk.JavaToRuntimeID("minecraft:mangrove_propagule[age=0,hanging=true,stage=0,waterlogged=true]")
	// blockName, state, _ = chunk.RuntimeIDToState(rtid)
	// fmt.Println(blockName, state)
	rtid, blockName, state, blockNameWithStateStr, StateStr := uint32(0), "", map[string]any{}, "", ""
	rtid, _ = blocks.LegacyBlockToRuntimeID("cartography_table", 0)
	fmt.Println(rtid)
	fmt.Println(blocks.RuntimeIDToBlock(rtid))

	rtid, _ = blocks.LegacyBlockToRuntimeID("coral_block", 4)
	fmt.Println(rtid)
	blockName, state, _ = blocks.RuntimeIDToState(rtid)
	fmt.Println(blockName, state)
	blockNameWithStateStr, _ = blocks.RuntimeIDToBlockNameWithStateStr(rtid)
	fmt.Println(blockNameWithStateStr)
	blockName, StateStr, _ = blocks.RuntimeIDToBlockNameAndStateStr(rtid)
	fmt.Println(blockName, StateStr)
	fmt.Println(blocks.BlockNameAndStateToRuntimeID(blockName, state))
	fmt.Println(blocks.BlockNameAndStateStrToRuntimeID(blockName, StateStr))

	rtid, _ = blocks.BlockStrToRuntimeID("cartography_table")
	fmt.Println(blocks.RuntimeIDToBlockNameWithStateStr(rtid))
	rtid, _ = blocks.BlockStrToRuntimeID(`coral_block ["coral_color":"yellow", "dead_bit":false]`)
	fmt.Println(blocks.RuntimeIDToBlockNameWithStateStr(rtid))
	rtid, _ = blocks.BlockStrToRuntimeID(`coral_block["coral_color":"yellow", "dead_bit":false]`)
	fmt.Println(blocks.RuntimeIDToBlockNameWithStateStr(rtid))
	rtid, _ = blocks.BlockStrToRuntimeID(`coral_block["coral_color":"y"]`)
	fmt.Println(blocks.RuntimeIDToBlockNameWithStateStr(rtid))
	rtid, _ = blocks.BlockStrToRuntimeID(`minecraft:sapling`)
	fmt.Println(blocks.RuntimeIDToBlockNameWithStateStr(rtid))
	rtid, _ = blocks.BlockStrToRuntimeID(`minecraft:dark_oak_sapling[stage=1]`)
	fmt.Println(blocks.RuntimeIDToBlockNameWithStateStr(rtid))
	rtid, _ = blocks.BlockStrToRuntimeID(`dark_oak_sapling[stage=1]`)
	fmt.Println(blocks.RuntimeIDToBlockNameWithStateStr(rtid))
	rtid, _ = blocks.BlockStrToRuntimeID(`minecraft:mangrove_propagule[age=0,hanging=true,stage=0,waterlogged=true]`)
	fmt.Println(blocks.RuntimeIDToBlockNameWithStateStr(rtid))
	rtid = blocks.SchematicToRuntimeID(35, 3)
	fmt.Println(blocks.RuntimeIDToBlockNameWithStateStr(rtid))
	rtid, ok := blocks.BlockStrToRuntimeID(`wool ["color":"orange"]`)
	if !ok {
		panic("not ok")
	}
	fmt.Println(blocks.RuntimeIDToBlockNameAndStateStr(rtid))
	rtid, ok = blocks.BlockStrToRuntimeID(`stone   2`)
	if !ok {
		panic("not ok")
	}
	fmt.Println(blocks.RuntimeIDToBlockNameAndStateStr(rtid))
	rtid, ok = blocks.BlockStrToRuntimeID(`grass`)
	if !ok {
		panic("not ok")
	}
	fmt.Println(blocks.RuntimeIDToBlockNameAndStateStr(rtid))
	rtid, ok = blocks.SchemBlockStrToRuntimeID(`grass`)
	if !ok {
		panic("not ok")
	}
	fmt.Println(blocks.RuntimeIDToBlockNameAndStateStr(rtid))
	rtid, ok = blocks.SchemBlockStrToRuntimeID(`wool 4`)
	if !ok {
		panic("not ok")
	}
	fmt.Println(blocks.RuntimeIDToBlockNameAndStateStr(rtid))
	rtid, ok = blocks.SchemBlockStrToRuntimeID(`shulker_box 3`)
	if !ok {
		panic("not ok")
	}
	fmt.Println(blocks.RuntimeIDToBlockNameAndStateStr(rtid))
}
