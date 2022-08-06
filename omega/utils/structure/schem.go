package structure

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"reflect"

	"github.com/Tnze/go-mc/nbt"
)

type SchemFileStructrue struct {
	Palette       map[string]int32
	Metadata      map[string]interface{}
	DataVersion   int32
	BlockDataIn   []byte      `nbt:"BlockData"`
	OffsetIn      interface{} `nbt:"Offset"`
	blockData     []uint16
	offset        []int
	PaletteMax    int32
	Version       int32
	Length        int16
	Height        int16
	Width         int16
	BlockEntities []map[string]interface{}
}

func DecodeSchem(data []byte, infoSender func(string)) (blockFeeder chan *IOBlock, cancelFn func(), suggestMinCacheChunks int, totalBlocks int, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("unknown error %v", r)
		}
	}()
	err = ErrImportFormateNotSupport
	hasGzipHeader := false
	if len(data) < 2 {
		return nil, nil, 0, 0, err
	}
	if data[0] == 0x1f && data[1] == 0x8b {
		hasGzipHeader = true
	}
	var dataFeeder io.Reader
	if !hasGzipHeader {
		dataFeeder = bytes.NewBuffer(data)
	} else {
		dataFeeder, err = gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, nil, 0, 0, err
		}
	}
	nbtDecoder := nbt.NewDecoder(dataFeeder)
	var schemData SchemFileStructrue
	infoSender("解压缩数据，将消耗大量内存")
	_, err = nbtDecoder.Decode(&schemData)
	infoSender("解压缩成功")
	if err != nil || schemData.BlockDataIn == nil || len(schemData.BlockDataIn) == 0 {
		return nil, nil, 0, 0, ErrImportFormateNotSupport
	}

	// convertBlockData := reflect.ValueOf(schemData.BlockDataIn)
	// dataLen := convertBlockData.Len()
	blockLen := 0
	schemData.blockData = make([]uint16, len(schemData.BlockDataIn))

	currOffset := uint8(0)
	currLowByte := uint16(0)
	for i := 0; i < len(schemData.BlockDataIn); i++ {
		currentByte := schemData.BlockDataIn[i]
		if currentByte&128 != 0 {
			currLowByte |= uint16((currentByte)&127) << (currOffset * 7)
			currOffset++
			continue
		}
		if currOffset == 0 {
			blockLen++
			schemData.blockData[blockLen] = uint16(currentByte)
		} else {
			blockLen++
			currLowByte |= uint16((currentByte)&127) << (currOffset * 7)
			schemData.blockData[blockLen] = currLowByte
			currLowByte = 0
			currOffset = 0
		}
	}
	schemData.blockData = schemData.blockData[:blockLen]
	schemData.BlockDataIn = nil
	convertOffset := reflect.ValueOf(schemData.OffsetIn)
	dataLen := convertOffset.Len()
	schemData.offset = make([]int, dataLen)
	for i := 0; i < dataLen; i++ {
		rv := convertOffset.Index(i)
		if !rv.CanInt() {
			return nil, nil, 0, 0, fmt.Errorf("cannot convert offset %v to int, please contact 2401PT", rv)
		}
		schemData.offset[i] = int(rv.Int())
	}
	paletteMapping := map[int32]uint32{}
	for javaBlockStr, paletteI := range schemData.Palette {
		RTID, found := chunk.JavaToRuntimeID(javaBlockStr)
		if !found {
			infoSender(fmt.Sprintf("does not know what to do with %v, so handle it as air", javaBlockStr))
		} else {
			// infoSender(fmt.Sprintf("found"))
		}
		paletteMapping[paletteI] = RTID
	}
	fmt.Printf("schem file size %v %v %v\n", schemData.Width, schemData.Height, schemData.Length)
	if len(schemData.blockData) != int(schemData.Height)*int(schemData.Width)*int(schemData.Length) {
		return nil, nil, 0, 0, fmt.Errorf("size check fail %v != %v", schemData.Width*schemData.Height*schemData.Length, len(schemData.blockData))
	}

	Nbts := make(map[define.CubePos]map[string]interface{})
	for _, nbt := range schemData.BlockEntities {
		pos := nbt["Pos"].([]int32)
		// fmt.Println(nbt)
		Nbts[define.CubePos{int(pos[0]), int(pos[1]), int(pos[2])}] = nbt
	}

	width, height, length := int(schemData.Width), int(schemData.Height), int(schemData.Length)
	blockData := schemData.blockData
	blockChan := make(chan *IOBlock, 10240)
	stop := false
	airRID := chunk.AirRID
	blocksCounter := 0
	for _, blk := range blockData {
		if paletteMapping[int32(blk)] != airRID {
			blocksCounter++
		}
	}
	infoSender(fmt.Sprintf("格式匹配成功,开始解析,尺寸 [%v, %v, %v] 方块数量 %v\n", width, height, length))
	go func() {
		defer func() {
			close(blockChan)
		}()
		x, y, z, index := 0, 0, 0, 0
		blkRTID := uint32(0)
		var p define.CubePos
		for z = 0; z < length; z++ {
			for y = 0; y < height; y++ {
				for x = 0; x < width; x++ {
					if stop {
						return
					}
					index = x + z*width + y*length*width
					blkRTID = paletteMapping[int32(blockData[index])]
					if blkRTID == airRID {
						continue
					}
					p = define.CubePos{x, y, z}
					blockChan <- &IOBlock{Pos: p, RTID: blkRTID, NBT: Nbts[p]}
				}
			}
		}
	}()
	return blockChan, func() {
		stop = true
	}, (suggestMinCacheChunks / 16) + 1, blocksCounter, nil
}
