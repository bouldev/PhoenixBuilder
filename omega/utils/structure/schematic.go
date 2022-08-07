package structure

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"

	"github.com/Tnze/go-mc/nbt"
	"github.com/pterm/pterm"
)

func DecodeSchematic(data []byte, infoSender func(string)) (blockFeeder chan *IOBlock, cancelFn func(), suggestMinCacheChunks int, totalBlocks int, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("unknown error %v", r)
		}
	}()
	err = ErrImportFormateNotSupport
	var dataFeeder io.Reader
	dataFeeder, err = gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		fmt.Println("fail in gzip")
		return nil, nil, 0, 0, err
	}

	nbtDecoder := nbt.NewDecoder(dataFeeder)
	var schematicData struct {
		Blocks           []byte `nbt:"Blocks"`
		Data             []byte `nbt:"Data"`
		Width            int16  `nbt:"Width"`
		Length           int16  `nbt:"Length"`
		Height           int16  `nbt:"Height"`
		WEOffsetX        int    `nbt:"WEOffsetX"`
		WEOffsetY        int    `nbt:"WEOffsetY"`
		WEOffsetZ        int    `nbt:"WEOffsetZ"`
		Materials        string
		ItemStackVersion uint8 `nbt:"itemStackVersion"`
	}
	infoSender("解压缩数据，将消耗大量内存")
	_, err = nbtDecoder.Decode(&schematicData)
	infoSender("解压缩成功")
	if err != nil {
		// fmt.Println("fail in formate check", err, schematicData)
		return nil, nil, 0, 0, ErrImportFormateNotSupport
	}
	blocks := schematicData.Blocks
	values := schematicData.Data
	if schematicData.Blocks == nil || len(blocks) == 0 || schematicData.Data == nil || len(values) == 0 {
		// fmt.Println("fail in formate check", err, schematicData)
		return nil, nil, 0, 0, ErrImportFormateNotSupport
	}
	Size := [3]int{int(schematicData.Width), int(schematicData.Height), int(schematicData.Length)}
	X, Y, Z := 0, 1, 2
	// fmt.Printf("schematic file size %v %v %v\n", Size[X], Size[Y], Size[Z])
	if len(blocks) != int(Size[X])*int(Size[Y])*int(Size[Z]) {
		return nil, nil, 0, 0, fmt.Errorf("size check fail %v != %v", int(Size[X])*int(Size[Y])*int(Size[Z]), len(blocks))
	}
	blockChan := make(chan *IOBlock, 10240)
	airRID := chunk.AirRID
	lightBlockRID, found := chunk.LegacyBlockToRuntimeID("light_block", 0)
	if !found {
		pterm.Error.Println("placeholder block RTID not found")
	}
	blocksCounter := 0
	for _, blk := range blocks {
		if blk != 0 {
			blocksCounter++
		}
	}
	stop := false
	infoSender(fmt.Sprintf("格式匹配成功,开始解析,尺寸 %v, 方块数量 %v\n", Size, blocksCounter))
	go func() {
		defer func() {
			close(blockChan)
		}()
		width, height, length := Size[X], Size[Y], Size[Z]
		index, name, data := 0, "", uint8(0)
		rtid, found := uint32(0), false
		x, y, z := 0, 0, 0
		blkSchematicID := byte(0)
		notFoundCache := map[string]bool{}
		for z = 0; z < length; z++ {
			for y = 0; y < height; y++ {
				for x = 0; x < width; x++ {
					if stop {
						return
					}
					index = x + z*width + y*length*width
					blkSchematicID = blocks[index]
					if blkSchematicID == 0 {
						continue
					}
					name = chunk.SchematicBlockMapping[blkSchematicID]
					data = uint8(values[index])
					rtid, found = chunk.LegacyBlockToRuntimeID(name, data)
					if !found {
						rtid, found = chunk.LegacyBlockToRuntimeID(name, 0)
						if !found {
							if _, hasK := notFoundCache[name]; !hasK {
								infoSender(fmt.Sprintf("Warning: %v not support in Schematic Foramte", name))
								notFoundCache[name] = true
							}
						}
						// continue
					} else {
						// fmt.Printf("%v,%v,%v\t", name, data, rtid)
					}
					if rtid != airRID {
						blockChan <- &IOBlock{Pos: define.CubePos{x, y, z}, RTID: rtid}
					} else {
						blockChan <- &IOBlock{Pos: define.CubePos{x, y, z}, RTID: lightBlockRID}
					}
				}
			}
		}
	}()
	return blockChan, func() {
		stop = true
	}, (suggestMinCacheChunks / 16) + 1, blocksCounter, nil
}
