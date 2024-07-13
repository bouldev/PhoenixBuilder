package tiled_buffer

import (
	"bytes"
	"fmt"

	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/block_actors"

	"github.com/mitchellh/mapstructure"
)

// 将类型为 ID 的方块实体的 __tag NBT 数据从 buffer 底层输出流解码
func Decode(ID string, buffer *bytes.Buffer) (block_actors.BlockActors, error) {
	reader := protocol.NewReader(buffer, 0, false)
	block, success := block_actors.NewPool()[ID]
	if !success {
		return nil, fmt.Errorf("Decode: Can not get target block NBT method; ID = %#v", ID)
	}
	block.Marshal(reader)
	return block, nil
}

// 将 block 编码为 __tag NBT 的二进制数据，
// 同时返回该方块实体对应的 ID 名
func Encode(block block_actors.BlockActors) (ID string, bytesGet []byte) {
	buffer := bytes.NewBuffer([]byte{})
	writer := protocol.NewWriter(buffer, 0)
	block.Marshal(writer)
	return block.ID(), buffer.Bytes()
}

// 将 block 写入到一个空切片中，
// 然后从该切片重新阅读数据，
// 并返回该方块实体对应的 NBT 表达形式。
// ID 是该方块实体的 ID 名
func WriteAndRead(ID string, block block_actors.BlockActors) (map[string]any, error) {
	var mapping map[string]any
	// prepare
	id, blockBytes := Encode(block)
	if id != ID {
		return nil, fmt.Errorf("WriteAndRead: ID of block NBT is not matched; id = %#v, ID = %#v", id, ID)
	}
	// write
	buffer := bytes.NewBuffer(blockBytes)
	new, err := Decode(ID, buffer)
	if err != nil {
		return nil, fmt.Errorf("WriteAndRead: %v", err)
	}
	// read again
	if length := buffer.Len(); length > 0 {
		return nil, fmt.Errorf("WriteAndRead: %T: %v unread bytes left: 0x%x", block, length, buffer.Bytes())
	}
	// check unread parts
	if err = mapstructure.Decode(new, &mapping); err != nil {
		return nil, fmt.Errorf("WriteAndRead: %v", err)
	}
	// get nbt mapping
	return mapping, nil
	// return
}
