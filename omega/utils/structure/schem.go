package structure

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/memory"
	"phoenixbuilder/mirror/io/world"

	standard_nbt "phoenixbuilder/minecraft/nbt"

	"github.com/Tnze/go-mc/nbt"
)

type WEOffset struct {
	WEOffsetX int32
	WEOffsetY int32
	WEOffsetZ int32
}

type NbtBlocks []map[string]interface{}

func (n NbtBlocks) TagType() byte {
	return nbt.TagList
}

func (n NbtBlocks) Encode(w io.Writer) error {
	buf := bytes.NewBuffer([]byte{})
	if err := standard_nbt.NewEncoderWithEncoding(buf, standard_nbt.BigEndian).Encode(n); err != nil {
		return err
	} else {
		_, err := w.Write(buf.Bytes()[3:])
		return err
	}
}

type SchemFileStructrue struct {
	Palette       map[string]int32 // schemData.Palette["minecraft:ice"]=0
	Metadata      WEOffset         //map[string]interface {} ["WEOffsetX": -166 (int32), "WEOffsetY": -106 (int32), "WEOffsetZ": 25 (int32), ], WE=World Editor? set to the value of offset should be ok (y=height+offset)
	DataVersion   int32            //2975
	BlockDataIn   []byte           `nbt:"BlockData"` // var uint16 of  blockData
	Offset        []int32          // should be the structure pos in the origin wold
	blockData     []uint16         // uint16
	PaletteMax    int32            // len(Palette)
	Version       int32            // 2
	Length        int16
	Height        int16
	Width         int16 //x
	BlockEntities NbtBlocks
}

func convertNbtPos(nbt map[string]interface{}, sx, sy, sz int32) (pos [3]int32, success bool) {
	if v, found := nbt["x"]; found {
		delete(nbt, "x")
		pos[0] = v.(int32) - sx
	} else {
		return pos, false
	}
	if v, found := nbt["y"]; found {
		delete(nbt, "y")
		pos[1] = v.(int32) - sy
	} else {
		return pos, false
	}
	if v, found := nbt["z"]; found {
		delete(nbt, "z")
		pos[2] = v.(int32) - sz
	} else {
		return pos, false
	}
	return pos, true
}

func EncodeSchem(chunks map[define.ChunkPos]*mirror.ChunkData, startPos, endPos define.CubePos, structureName string, targetDir string) (err error) {
	schemFile := &SchemFileStructrue{
		Palette: make(map[string]int32),
		Metadata: WEOffset{
			WEOffsetX: 0,
			WEOffsetY: 0,
			WEOffsetZ: 0,
		},
		DataVersion:   2975,
		Offset:        []int32{0, 0, 0},
		PaletteMax:    0,
		Version:       2,
		Length:        int16(endPos.Z()) - int16(startPos.Z()),
		Height:        int16(endPos.Y()) - int16(startPos.Y()),
		Width:         int16(endPos.X()) - int16(startPos.X()),
		BlockEntities: NbtBlocks{},
	}
	numBlocks := (uint64(endPos.X()-startPos.X()) * uint64(endPos.Y()-startPos.Y()) * uint64(endPos.Z()-startPos.Z()))
	if numBlocks >= uint64(uint64(1)<<uint64(31)) {
		return fmt.Errorf("too many blocks!")
	}
	memoryHolder := memory.NewMemoryChunkCacher(chunks)
	srcWorld := world.NewWorld(memoryHolder)
	sx, sy, sz, ex, ey, ez := startPos.X(), startPos.Y(), startPos.Z(), endPos.X(), endPos.Y(), endPos.Z()
	cx, cy, cz := 0, 0, 0
	// ox, oy, oz := 0, 0, 0
	x8, y8, z8 := 0, 0, 0
	rtid, blockNbt := chunk.AirRID, make(map[string]interface{})
	paletteI := uint32(0)
	BlockEntities := NbtBlocks{}
	writerBlocks := bytes.NewBuffer(make([]byte, 0, uint32(float64(numBlocks)*1.3)))

	//	if currentByte&128 != 0 {
	// 	currLowByte |= uint16((currentByte)&127) << (currOffset * 7)
	// 	currOffset++
	// 	continue
	// }
	// if currOffset == 0 {
	// 	schemData.blockData[blockLen] = uint16(currentByte)
	// 	blockLen++
	// } else {
	// 	currLowByte |= uint16((currentByte)&127) << (currOffset * 7)
	// 	schemData.blockData[blockLen] = currLowByte
	// 	blockLen++
	// 	currLowByte = 0
	// 	currOffset = 0
	// }

	writeVarUint32 := func(u uint32) {
		for u&128 != 0 {
			_ = writerBlocks.WriteByte(byte(u) | 128)
			u >>= 7
		}
		_ = writerBlocks.WriteByte(byte(u))
	}
	convertor := NewRuntimeIDToPaletteConvertor()
	convertor.AcquirePaletteFN = func(u uint32) string {
		javaStr, _ := chunk.RuntimeIDToJava(u)
		return javaStr
	}
	convertor.Convert(chunk.AirRID) // force air to be 0
	for y8 = sy; y8 < ey; y8 += 8 {
		for cy = y8; cy < y8+8; cy++ {
			if cy == ey {
				break
			}
			// oy = cy - sy
			for z8 = sz; z8 < ez; z8 += 8 {
				for cz = z8; cz < z8+8; cz++ {
					if cz == ez {
						break
					}
					// oz = cz - sz
					for x8 = sx; x8 < ex; x8 += 8 {
						for cx = x8; cx < x8+8; cx++ {
							if cx == ex {
								break
							}
							// ox = cx - sx
							rtid, blockNbt, _ = srcWorld.BlockWithNbt(define.CubePos{cx, cy, cz})
							paletteI = convertor.Convert(rtid)
							writeVarUint32(paletteI)
							if blockNbt != nil {
								pos, success := convertNbtPos(blockNbt, int32(sx), int32(sy), int32(sz))
								if success {
									blockNbt["Pos"] = pos
									// delete(blockNbt, "id")
									// name := convertor.Palette[paletteI]
									// nameSplit := strings.Split(name, "[")
									// if len(nameSplit) > 1 {
									// 	blockNbt["Id"] = nameSplit[0]

									if blockNbt["id"] == "CommandBlock" {
										blockNbt["Id"] = "minecraft:command_block"
									}
									BlockEntities = append(BlockEntities, blockNbt)
									// }
								}
							}
						}
					}
				}
			}
		}
	}
	schemFile.BlockDataIn = writerBlocks.Bytes()
	for paletteI, paletteStr := range convertor.Palette {
		if repeatedPaletteI, found := schemFile.Palette[paletteStr]; found {
			return fmt.Errorf("repeated palette %v %v", repeatedPaletteI, paletteStr)
		}
		schemFile.Palette[paletteStr] = int32(paletteI)
	}
	schemFile.PaletteMax = int32(len(schemFile.Palette))
	schemFile.BlockEntities = BlockEntities
	fp, err := os.OpenFile(path.Join(targetDir, structureName+".schem"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer fp.Close()
	writer := gzip.NewWriter(fp)
	defer writer.Close()
	// nbtEncoder := standard_nbt.NewEncoderWithEncoding(writer, standard_nbt.BigEndian)
	// err = nbtEncoder.EncodeWithRootTag(*schemFile, "Schematic")
	err = nbt.NewEncoder(writer).Encode(*schemFile, "Schematic")
	if err != nil {
		return err
	}
	return nil
}

func DecodeSchem(data []byte, infoSender func(string)) (blockFeeder chan *IOBlockForDecoder, cancelFn func(), suggestMinCacheChunks int, totalBlocks int, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("unknown error %v", r)
		}
	}()
	err = ErrImportFormatNotSupport
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
	rootTag, err := nbtDecoder.Decode(&schemData)
	if rootTag != "Schematic" || err != nil || schemData.BlockDataIn == nil || len(schemData.BlockDataIn) == 0 {
		return nil, nil, 0, 0, ErrImportFormatNotSupport
	}
	infoSender("解压缩成功")

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
			schemData.blockData[blockLen] = uint16(currentByte)
			blockLen++
		} else {
			currLowByte |= uint16((currentByte)&127) << (currOffset * 7)
			schemData.blockData[blockLen] = currLowByte
			blockLen++
			currLowByte = 0
			currOffset = 0
		}
	}
	schemData.blockData = schemData.blockData[:blockLen]
	schemData.BlockDataIn = nil
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
		return nil, nil, 0, 0, fmt.Errorf("size check fail %v * %v * %v = %v != %v", schemData.Width, schemData.Height, schemData.Length, int(schemData.Height)*int(schemData.Width)*int(schemData.Length), len(schemData.blockData))
	}

	Nbts := make(map[define.CubePos]map[string]interface{})
	for _, nbt := range schemData.BlockEntities {
		pos := nbt["Pos"].([]int32)
		// fmt.Println(nbt)
		Nbts[define.CubePos{int(pos[0]), int(pos[1]), int(pos[2])}] = nbt
	}

	width, height, length := int(schemData.Width), int(schemData.Height), int(schemData.Length)
	blockData := schemData.blockData
	blockChan := make(chan *IOBlockForDecoder, 10240)
	stop := false
	airRID := chunk.AirRID
	blocksCounter := 0
	for _, blk := range blockData {
		if paletteMapping[int32(blk)] != airRID {
			blocksCounter++
		}
	}
	infoSender(fmt.Sprintf("格式匹配成功,开始解析,尺寸 [%v, %v, %v] 方块数量 %v\n", width, height, length, blocksCounter))
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
					blockChan <- &IOBlockForDecoder{Pos: p, RTID: blkRTID, NBT: Nbts[p]}
				}
			}
		}
	}()
	return blockChan, func() {
		stop = true
	}, (width / 16) + 2, blocksCounter, nil
}
