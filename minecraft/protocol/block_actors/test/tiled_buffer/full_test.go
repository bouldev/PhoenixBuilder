package tiled_buffer

import (
	"bytes"
	"reflect"
	"testing"

	"phoenixbuilder/minecraft/nbt"
	"phoenixbuilder/minecraft/protocol/block_actors"

	"github.com/mitchellh/mapstructure"
	"github.com/pterm/pterm"
)

// 测试 __tag NBT <-> NBT Go Struct <-> NBT Map 是否正常工作
func TestFull(t *testing.T) {
	for _, element := range NewPool() {
		var blockNBTMap map[string]any
		// prepare
		{
			buffer := bytes.NewBuffer(element.Buffer)
			block, err := Decode(element.ID, buffer)
			if err != nil {
				t.Errorf("TestFull: %v", err)
			}
			if err = mapstructure.Decode(block, &blockNBTMap); err != nil {
				t.Errorf("TestFull: %v", err)
			}
			// read
			if length := buffer.Len(); length > 0 {
				t.Errorf("%T: %v unread bytes left: 0x%x", block, length, buffer.Bytes())
			}
			// check unread parts
			secondBlockMap, err := WriteAndRead(element.ID, block)
			if err != nil {
				t.Errorf("TestFull: %v", err)
			}
			if !reflect.DeepEqual(blockNBTMap, secondBlockMap) {
				t.Errorf("TestFull: Marshal and unmarshal is unequivalence; element.ID = %#v", element.ID)
			}
			// write
		}
		// __tag NBT <-> NBT Go Struct
		{
			var newBlockNBTMap map[string]any
			new := block_actors.NewPool()[element.ID]
			// prepare
			err := mapstructure.Decode(blockNBTMap, &new)
			if err != nil {
				t.Errorf("TestFull: %v", err)
			}
			if err := mapstructure.Decode(new, &newBlockNBTMap); err != nil {
				t.Errorf("TestFull: %v", err)
			}
			// to nbt
			if !reflect.DeepEqual(blockNBTMap, newBlockNBTMap) {
				t.Errorf("TestFull: NBT Map convert is unequivalence; element.ID = %#v", element.ID)
			}
		}
		// NBT Go Struct <-> NBT Map
		if _, err := nbt.MarshalEncoding(blockNBTMap, nbt.LittleEndian); err != nil {
			t.Errorf("TestFull: %v", err)
		}
		// check NBT encode
		pterm.Success.Printf("%#v\n", blockNBTMap)
		// print success
	}
}
