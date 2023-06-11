package structure

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/world_provider"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"

	"github.com/andybalholm/brotli"
)

func ReadBrString(src io.Reader) (string, error) {
	str := ""
	c := make([]byte, 1)
	for {
		_, err := src.Read(c)
		if err != nil {
			return "", err
		}
		if c[0] == 0 {
			break
		}
		str += string(c)
	}
	return str, nil
}

func handleBDXCMD(br io.Reader, infoSender func(string)) (author string, blockChan chan *IOBlockForDecoder, cancelFn func(), err error) {
	legacyRunTimeIDRemapper := NewDoubleValueLegacyBlockToRuntimeIDMapper()
	blockChan = make(chan *IOBlockForDecoder, 10240)
	{
		tempbuf := make([]byte, 4)
		_, err = br.Read(tempbuf)
		if err != nil {
			err = fmt.Errorf(I18n.T(I18n.InvalidFileError))
			return
		}
		if string(tempbuf) != "BDX\x00" {
			err = fmt.Errorf(I18n.T(I18n.BDump_NotBDX_Invinnerheader))
			return
		}
	}
	author, _err := ReadBrString(br)
	if _err != nil {
		err = fmt.Errorf("cannot read author info, %v", err)
		return
	}
	curcmdbuf := make([]byte, 1)
	brushPosition := define.CubePos{0, 0, 0}
	paletteIDCounter := uint16(0)
	runtimeIdPoolUsing := NewRuntimeIDConvertor()
	stop := false
	go func() {
		defer close(blockChan)
		cmd := uint8(0)
		for {
			if brushPosition[1] > 200 {
				fmt.Println("error!", cmd)
			}
			if stop {
				break
			}
			n, err := br.Read(curcmdbuf)
			if err != nil || n != 1 {
				infoSender("Failed to get construction command, file may be corrupted")
				return
			}
			cmd = curcmdbuf[0]
			if cmd == 88 {
				return
			}
			if cmd == 1 {
				bstr, _err := ReadBrString(br)
				if _err != nil {
					infoSender("Failed to get argument for cmd[pos:0], file may be corrupted!")
					return
				}
				legacyRunTimeIDRemapper.AddBlockNamePalette(paletteIDCounter, bstr)
				paletteIDCounter++
				continue
			} else if cmd == 2 {
				rdst := make([]byte, 2)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos1], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint16(rdst)
				brushPosition[0] += int(jumpval)
				brushPosition[1] = 0
				brushPosition[2] = 0
			} else if cmd == 3 {
				brushPosition[0]++
				brushPosition[1] = 0
				brushPosition[2] = 0
			} else if cmd == 4 {
				rdst := make([]byte, 2)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos2], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint16(rdst)
				brushPosition[1] += int(jumpval)
				brushPosition[2] = 0
			} else if cmd == 5 {
				brushPosition[1]++
				brushPosition[2] = 0
			} else if cmd == 6 {
				rdst := make([]byte, 2)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos3], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint16(rdst)
				brushPosition[2] += int(jumpval)
			} else if cmd == 7 {
				rdst := make([]byte, 2)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos4], file may be corrupted")
					return
				}
				blockId := binary.BigEndian.Uint16(rdst)
				blockDataBytes := make([]byte, 2)
				_, err = br.Read(blockDataBytes)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos5], file may be corrupted")
					return
				}
				blockData := binary.BigEndian.Uint16(blockDataBytes)
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: legacyRunTimeIDRemapper.GetRTID(blockId, blockData),
				}
			} else if cmd == 8 {
				brushPosition[2]++
			} else if cmd == 9 {
				// Command: NOP
			} else if cmd == 10 {
				rdst := make([]byte, 4)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos6], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint32(rdst)
				brushPosition[0] += int(jumpval)
				brushPosition[1] = 0
				brushPosition[2] = 0
			} else if cmd == 11 {
				rdst := make([]byte, 4)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos7], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint32(rdst)
				brushPosition[1] += int(jumpval)
				brushPosition[2] = 0
			} else if cmd == 12 {
				rdst := make([]byte, 4)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos8], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint32(rdst)
				brushPosition[2] += int(jumpval)
			} else if cmd == 13 {
				fmt.Printf("WARNING: BDump/Import: Use of reserved command\n")
			} else if cmd == 14 {
				brushPosition[0]++
			} else if cmd == 15 {
				brushPosition[0]--
			} else if cmd == 16 {
				brushPosition[1]++
			} else if cmd == 17 {
				brushPosition[1]--
			} else if cmd == 18 {
				brushPosition[2]++
			} else if cmd == 19 {
				brushPosition[2]--
			} else if cmd == 20 {
				rdst := make([]byte, 2)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos9], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint16(rdst)
				brushPosition[0] += int(int16(jumpval))
			} else if cmd == 21 {
				rdst := make([]byte, 4)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos10], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint32(rdst)
				brushPosition[0] += int(int32(jumpval))
			} else if cmd == 22 {
				rdst := make([]byte, 2)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos11], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint16(rdst)
				brushPosition[1] += int(int16(jumpval))
			} else if cmd == 23 {
				rdst := make([]byte, 4)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos12], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint32(rdst)
				brushPosition[1] += int(int32(jumpval))
			} else if cmd == 24 {
				rdst := make([]byte, 2)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos13], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint16(rdst)
				brushPosition[2] += int(int16(jumpval))
			} else if cmd == 25 {
				rdst := make([]byte, 4)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos14], file may be corrupted")
					return
				}
				jumpval := binary.BigEndian.Uint32(rdst)
				brushPosition[2] += int(int32(jumpval))
			} else if cmd == 26 {
				fbuf := make([]byte, 4)
				_, err := br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos15], file may be corrupted")
					return
				}
				cbmode := binary.BigEndian.Uint32(fbuf)
				command, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos16], file may be corrupted")
					return
				}
				cusname, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos17], file may be corrupted")
					return
				}
				lasout, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos18], file may be corrupted")
					return
				}
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos19], file may be corrupted")
					return
				}
				tickdelay := int32(binary.BigEndian.Uint32(fbuf))
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos20], file may be corrupted")
					return
				}
				blockName := "command_block"
				switch cbmode {
				case packet.CommandBlockImpulse:
				case packet.CommandBlockRepeating:
					blockName = "repeating_command_block"
				case packet.CommandBlockChain:
					blockName = "chain_command_block"
				}
				rtid, _ := chunk.LegacyBlockToRuntimeID(blockName, 0)
				commandBlockNbt := map[string]interface{}{
					"id":                 "CommandBlock",
					"Command":            string(command),
					"CustomName":         string(cusname),
					"ExecuteOnFirstTick": uint8(fbuf[0]),
					"TickDelay":          int32(tickdelay),
					"auto":               1 - uint8(fbuf[3]),
					"TrackOutput":        uint8(fbuf[1]),
					"LastOutput":         string(lasout),
					"conditionalMode":    uint8(fbuf[2]),
				}
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: rtid,
					NBT:  commandBlockNbt,
				}
			} else if cmd == 27 {
				rdst := make([]byte, 2)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos4], file may be corrupted")
					return
				}
				blockId := binary.BigEndian.Uint16(rdst)
				_, err = br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos5], file may be corrupted")
					return
				}
				blockData := binary.BigEndian.Uint16(rdst)
				rtid := legacyRunTimeIDRemapper.GetRTID(blockId, blockData)
				fbuf := make([]byte, 4)
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos15], file may be corrupted")
					return
				}
				_ = binary.BigEndian.Uint32(fbuf)
				command, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos16], file may be corrupted")
					return
				}
				cusname, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos17], file may be corrupted")
					return
				}
				lasout, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos18], file may be corrupted")
					return
				}
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos19], file may be corrupted")
					return
				}
				tickdelay := int32(binary.BigEndian.Uint32(fbuf))
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos20], file may be corrupted")
					return
				}
				commandBlockNbt := map[string]interface{}{
					"id":                 "CommandBlock",
					"Command":            string(command),
					"CustomName":         string(cusname),
					"ExecuteOnFirstTick": uint8(fbuf[0]),
					"TickDelay":          int32(tickdelay),
					"auto":               1 - uint8(fbuf[3]),
					"TrackOutput":        uint8(fbuf[1]),
					"LastOutput":         string(lasout),
					"conditionalMode":    uint8(fbuf[2]),
				}
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: rtid,
					NBT:  commandBlockNbt,
				}
			} else if cmd == 36 {
				rdst := make([]byte, 2)
				_, err = br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos5], file may be corrupted")
					return
				}
				blockData := binary.BigEndian.Uint16(rdst)
				fbuf := make([]byte, 4)
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos15], file may be corrupted")
					return
				}
				cbmode := binary.BigEndian.Uint32(fbuf)
				command, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos16], file may be corrupted")
					return
				}
				cusname, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos17], file may be corrupted")
					return
				}
				lasout, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos18], file may be corrupted")
					return
				}
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos19], file may be corrupted")
					return
				}
				tickdelay := int32(binary.BigEndian.Uint32(fbuf))
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos20], file may be corrupted")
					return
				}
				blockName := "command_block"
				switch cbmode {
				case packet.CommandBlockImpulse:
				case packet.CommandBlockRepeating:
					blockName = "repeating_command_block"
				case packet.CommandBlockChain:
					blockName = "chain_command_block"
				}
				rtid, _ := chunk.LegacyBlockToRuntimeID(blockName, blockData)
				commandBlockNbt := map[string]interface{}{
					"id":                 "CommandBlock",
					"Command":            string(command),
					"CustomName":         string(cusname),
					"ExecuteOnFirstTick": uint8(fbuf[0]),
					"TickDelay":          int32(tickdelay),
					"auto":               1 - uint8(fbuf[3]),
					"TrackOutput":        uint8(fbuf[1]),
					"LastOutput":         string(lasout),
					"conditionalMode":    uint8(fbuf[2]),
				}
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: rtid,
					NBT:  commandBlockNbt,
				}
			} else if cmd == 28 {
				rdst := make([]byte, 1)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos21], file may be corrupted")
					return
				}
				brushPosition[0] += int(int8(rdst[0]))
			} else if cmd == 29 {
				rdst := make([]byte, 1)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos22], file may be corrupted")
					return
				}
				brushPosition[1] += int(int8(rdst[0]))
			} else if cmd == 30 {
				rdst := make([]byte, 1)
				_, err := br.Read(rdst)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos23], file may be corrupted")
					return
				}
				brushPosition[2] += int(int8(rdst[0]))
			} else if cmd == 31 {
				poolId := make([]byte, 1)
				_, err := br.Read(poolId)
				if err != nil {
					infoSender("Failed to get pool id, file may be corrupted.")
					return
				}
				var pool []*types.ConstBlock
				if poolId[0] == 117 {
					pool = world_provider.RuntimeIdArray_117
				} else if poolId[0] == 118 {
					pool = world_provider.RuntimeIdArray_2_1_10
				} else {
					infoSender("This file is using an unknown runtime id pool, we're unable to resolve it.")
					return
				}
				runtimeIdPoolUsing = NewRuntimeIDConvertor()
				runtimeIdPoolUsing.ConvertFN = func(u uint32) uint32 {
					blk := pool[u].Take()
					if rtid, found := chunk.LegacyBlockToRuntimeID(*blk.Name, blk.Data); found {
						return rtid
					} else {
						return chunk.AirRID
					}
				}
			} else if cmd == 32 {
				runtimeIdMem := make([]byte, 2)
				_, err = br.Read(runtimeIdMem)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos24], file may be corrupted")
					return
				}
				runtimeId := uint32(binary.BigEndian.Uint16(runtimeIdMem))
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: runtimeIdPoolUsing.Convert(runtimeId),
				}
			} else if cmd == 33 {
				runtimeIdMem := make([]byte, 4)
				_, err = br.Read(runtimeIdMem)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos24], file may be corrupted")
					return
				}
				runtimeId := binary.BigEndian.Uint32(runtimeIdMem)
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: runtimeIdPoolUsing.Convert(runtimeId),
				}
			} else if cmd == 34 || cmd == 35 {
				var dataval uint32
				if cmd == 34 {
					datavbuf := make([]byte, 2)
					_, err = br.Read(datavbuf)
					dataval = uint32(binary.BigEndian.Uint16(datavbuf))
				} else {
					datavbuf := make([]byte, 4)
					_, err = br.Read(datavbuf)
					dataval = binary.BigEndian.Uint32(datavbuf)
				}
				fbuf := make([]byte, 4)
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos a15], file may be corrupted")
					return
				}
				_ = binary.BigEndian.Uint32(fbuf)
				command, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos a16], file may be corrupted")
					return
				}
				cusname, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos a 17], file may be corrupted")
					return
				}
				lasout, err := ReadBrString(br)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos a18], file may be corrupted")
					return
				}
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos a19], file may be corrupted")
					return
				}
				tickdelay := int32(binary.BigEndian.Uint32(fbuf))
				_, err = br.Read(fbuf)
				if err != nil {
					infoSender("Failed to get argument for cmd[pos a20], file may be corrupted")
					return
				}
				commandBlockNbt := map[string]interface{}{
					"id":                 "CommandBlock",
					"Command":            string(command),
					"CustomName":         string(cusname),
					"ExecuteOnFirstTick": uint8(fbuf[0]),
					"TickDelay":          int32(tickdelay),
					"auto":               1 - uint8(fbuf[3]),
					"TrackOutput":        uint8(fbuf[1]),
					"LastOutput":         string(lasout),
					"conditionalMode":    uint8(fbuf[2]),
				}
				blockChan <- &IOBlockForDecoder{
					Pos:  brushPosition,
					RTID: runtimeIdPoolUsing.Convert(dataval),
					NBT:  commandBlockNbt,
				}
			} else if cmd == 37 || cmd == 38 {
				// var runtimeId uint32
				if cmd == 37 {
					rIdBuf := make([]byte, 2)
					_, err = br.Read(rIdBuf)
					// runtimeId = uint32(binary.BigEndian.Uint16(rIdBuf))
				} else {
					rIdBuf := make([]byte, 4)
					_, err = br.Read(rIdBuf)
					// runtimeId = binary.BigEndian.Uint32(rIdBuf)
				}
				slotCountCon := make([]byte, 1)
				br.Read(slotCountCon)
				// chest := make(types.ChestData, slotCountCon[0])
				for i := uint8(0); i < slotCountCon[0]; i++ {
					_, _ = ReadBrString(br)
					countcon := make([]byte, 1)
					damageBuf := make([]byte, 2)
					slotcon := make([]byte, 1)
					br.Read(countcon)
					br.Read(damageBuf)
					br.Read(slotcon)
					// damageVal := binary.BigEndian.Uint16(damageBuf)
					// chest[i] = types.ChestSlot{
					// 	Name:   itemname,
					// 	Count:  countcon[0],
					// 	Damage: damageVal,
					// 	Slot:   slotcon[0],
					// }
				}
				// pos := types.Position{
				// 	X: brushPosition[0] + config.Position.X,
				// 	Y: brushPosition[1] + config.Position.Y,
				// 	Z: brushPosition[2] + config.Position.Z,
				// }
				// blc <- &types.Module{
				// 	Block: runtimeIdPoolUsing[runtimeId].Take(),
				// 	Point: pos,
				// }
				// for _, slot := range chest {
				// 	slotcopy := types.ChestSlot(slot)
				// 	blc <- &types.Module{
				// 		ChestSlot: &slotcopy,
				// 		Point:     pos,
				// 	}
				// }
			} else {
				// fmt.Println("ERROR!")
				infoSender(fmt.Sprintf("unimplemented method found : %d", cmd))
				return
			}
		}
	}()
	return author, blockChan, func() { stop = true }, nil
}

func checkAndGetBDXHeader(src io.Reader) (err error) {
	header3bytes := make([]byte, 3)
	_, err = src.Read(header3bytes)
	if err != nil {
		return fmt.Errorf("file too short, %v", err)

	}
	if string(header3bytes) != "BD@" {
		return fmt.Errorf(I18n.T(I18n.BDump_NotBDX_Invheader))
	}
	return nil
}

func DecodeBDX(data []byte, infoSender func(string)) (blockFeeder chan *IOBlockForDecoder, cancelFn func(), suggestMinCacheChunks int, totalBlocks int, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("unknown error %v", r)
		}
	}()
	file := bytes.NewBuffer(data)
	err = checkAndGetBDXHeader(file)
	if err != nil {
		err = ErrImportFormatNotSupport
		return
	}

	bro := brotli.NewReader(file)
	br := &bytes.Buffer{}
	filelen, _ := br.ReadFrom(bro)
	if filelen == 0 {
		err = fmt.Errorf("cannnot use brotli for decompress")
		return
	}
	infoSender("正在检查 BDX 文件，需要消耗大量时间")
	author, blockChan, _, err := handleBDXCMD(br, func(s string) {
		err = fmt.Errorf(s)
	})
	if err != nil {
		return
	}
	blockCounter := 0
	b0 := <-blockChan
	minX := b0.Pos.X()
	maxX := minX
	for b := range blockChan {
		x := b.Pos.X()
		if x < minX {
			minX = x
		} else if x > maxX {
			maxX = x
		}
		blockCounter++
	}
	infoSender(fmt.Sprintf("文件检查完毕, 方块数 %v", blockCounter))
	if err == nil {
		infoSender("author: " + author)
		file = bytes.NewBuffer(data)
		checkAndGetBDXHeader(file)
		bro := brotli.NewReader(file)
		br := &bytes.Buffer{}
		_, _ = br.ReadFrom(bro)
		_, blockChan, stopFn, _ := handleBDXCMD(br, func(s string) {})
		return blockChan, stopFn, (maxX-minX)/16 + 2, blockCounter, nil
	}
	return
}
