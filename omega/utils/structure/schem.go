package structure

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"phoenixbuilder/minecraft/nbt"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"reflect"
)

type SchemFileStructrue struct {
	Palette       map[string]int32
	Metadata      map[string]interface{}
	DataVersion   int32
	BlockDataIn   interface{} `nbt:"BlockData"`
	OffsetIn      interface{} `nbt:"Offset"`
	blockData     []uint32
	offset        []int
	PaletteMax    int32
	Version       int32
	Length        int16
	Height        int16
	Width         int16
	BlockEntities []map[string]interface{}
}

func DecodeSchem(data []byte, infoSender func(string)) (blockFeeder chan *IOBlock, cancelFn func(), err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("unknown error %v", r)
		}
	}()
	err = ErrImportFormateNotSupport
	hasGzipHeader := false
	if len(data) < 2 {
		return nil, nil, err
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
			return nil, nil, err
		}
	}
	nbtDecoder := nbt.NewDecoderWithEncoding(dataFeeder, nbt.BigEndian)
	var schemData SchemFileStructrue
	err = nbtDecoder.Decode(&schemData)
	if err != nil {
		return nil, nil, ErrImportFormateNotSupport
	}

	convertBlockData := reflect.ValueOf(schemData.BlockDataIn)
	dataLen := convertBlockData.Len()
	blockLen := 0
	schemData.blockData = make([]uint32, dataLen)

	currOffset := uint8(0)
	currLowByte := uint32(0)
	for i := 0; i < dataLen; i++ {
		rv := convertBlockData.Index(i)
		if !rv.CanUint() {
			return nil, nil, fmt.Errorf("cannot convert block data %v to uint, please contact 2401PT", rv)
		}
		currentByte := uint8(rv.Uint())
		if currentByte&128 != 0 {
			currLowByte |= uint32((currentByte)&127) << (currOffset * 7)
			currOffset++
			continue
		}
		if currOffset == 0 {
			blockLen++
			schemData.blockData[i] = uint32(currentByte)
		} else {
			blockLen++
			currLowByte |= uint32((currentByte)&127) << (currOffset * 7)
			schemData.blockData[i] = currLowByte
			currLowByte = 0
			currOffset = 0
		}
	}
	schemData.blockData = schemData.blockData[:blockLen]

	convertOffset := reflect.ValueOf(schemData.OffsetIn)
	dataLen = convertOffset.Len()
	schemData.offset = make([]int, dataLen)
	for i := 0; i < dataLen; i++ {
		rv := convertOffset.Index(i)
		if !rv.CanInt() {
			return nil, nil, fmt.Errorf("cannot convert offset %v to int, please contact 2401PT", rv)
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
		return nil, nil, fmt.Errorf("schem file size check fail %v", len(schemData.blockData))
	}

	Nbts := make(map[define.CubePos]map[string]interface{})
	for _, nbt := range schemData.BlockEntities {
		pos := nbt["Pos"].([3]int32)
		// fmt.Println(nbt)
		Nbts[define.CubePos{int(pos[0]), int(pos[1]), int(pos[2])}] = nbt
	}

	width, height, length := int(schemData.Width), int(schemData.Height), int(schemData.Length)
	blockData := schemData.blockData
	blockChan := make(chan *IOBlock, 4096)
	stop := false
	go func() {
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				for z := 0; z < length; z++ {
					if stop {
						return
					}
					index := y*width*length + z*width + x
					blkRTID := blockData[index]
					p := define.CubePos{x, y, z}
					b := &IOBlock{Pos: p, RTID: blkRTID}
					b.NBT = Nbts[p]
					blockChan <- b
				}
			}
		}
	}()
	return blockChan, func() {
		stop = true
	}, nil
}
