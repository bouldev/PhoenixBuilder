package bdump

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"phoenixbuilder/fastbuilder/bdump/command"
	"phoenixbuilder/fastbuilder/types"

	"github.com/andybalholm/brotli"
)

type BDumpLegacy struct {
	Author string // Should be empty
	Blocks []*types.Module
}

// containerIndex {blockName: {blockData: RunTimeIdIn117}}
var containerIndex map[string]map[int]int = map[string]map[int]int{
	"blast_furnace":      {0: 659, 1: 660, 2: 661, 3: 662, 4: 663, 5: 664},
	"lit_blast_furnace":  {0: 5413, 1: 5414, 2: 5415, 3: 5416, 4: 5417, 5: 5418},
	"smoker":             {0: 6636, 1: 6637, 2: 6638, 3: 6639, 4: 6640, 5: 6641},
	"lit_smoker":         {0: 5432, 1: 5433, 2: 5434, 3: 5435, 4: 5436, 5: 5437},
	"furnace":            {0: 4813, 1: 4814, 2: 4815, 3: 4816, 4: 4817, 5: 4818},
	"lit_furnace":        {0: 5420, 1: 5421, 2: 5422, 3: 5423, 4: 5424, 5: 5425},
	"chest":              {0: 1083, 1: 1084, 2: 1085, 3: 1086, 4: 1087, 5: 1088},
	"barrel":             {0: 201, 1: 202, 2: 203, 3: 204, 4: 205, 5: 206, 11: 210, 12: 211, 13: 212},
	"trapped_chest":      {0: 7135, 1: 7136, 2: 7137, 3: 7138, 4: 7139, 5: 7140},
	"lectern":            {0: 5339, 1: 5340, 2: 5341, 3: 5342, 4: 5343, 5: 5344, 6: 5345, 7: 5346},
	"hopper":             {0: 5019, 1: 5020, 2: 5021, 3: 5022, 4: 5023, 5: 5024, 8: 5025, 9: 5026, 10: 5027, 11: 5028, 12: 5029, 13: 5030},
	"dispenser":          {0: 4436, 1: 4437, 2: 4438, 3: 4439, 4: 4440, 5: 4441, 8: 4442, 9: 4443, 10: 4444, 11: 4445, 12: 4446, 13: 4447},
	"dropper":            {0: 4535, 1: 4536, 2: 4537, 3: 4538, 4: 4539, 5: 4540, 8: 4541, 9: 4542, 10: 4543, 11: 4544, 12: 4545, 13: 4546},
	"cauldron":           {0: 952, 1: 953, 2: 954, 3: 955, 4: 956, 5: 957, 6: 958, 8: 959, 9: 960, 10: 961, 11: 962, 12: 963, 13: 964, 14: 965, 16: 966, 17: 967, 18: 968, 19: 969, 20: 970, 21: 971, 22: 972},
	"lava_cauldron":      {0: 5294, 1: 5295, 2: 5296, 3: 5297, 4: 5298, 5: 5299, 6: 5300, 8: 5301, 9: 5302, 10: 5303, 11: 5304, 12: 5305, 13: 5306, 14: 5307, 16: 5308, 17: 5309, 18: 5310, 19: 5311, 20: 5312, 21: 5313, 22: 5314},
	"jukebox":            {0: 5113},
	"brewing_stand":      {0: 847, 1: 848, 2: 849, 3: 850, 4: 851, 5: 852, 6: 853, 7: 854},
	"undyed_shulker_box": {0: 7218},
	"shulker_box":        {0: 6586, 1: 6587, 2: 6588, 3: 6589, 4: 6590, 5: 6591, 6: 6592, 7: 6593, 8: 6594, 9: 6595, 10: 6596, 11: 6597, 12: 6598, 13: 6599, 14: 6600, 15: 6601},
}

/*
// Important note: Things under this file is what currently used,
//                 and actually bdump.go isn't under use now.


if(i.cmd=="addToBlockPalette"){
	writebuf(1,1);
	writebuf(i.blockName+"\0");
}else if(i.cmd=="addX"){
	writebuf(2,1);
	writebuf(i.count,2);
}else if(i.cmd=="X++"){
	writebuf(3,1);
}else if(i.cmd=="addY"){
	writebuf(4,1);
	writebuf(i.count,2);
}else if(i.cmd=="Y++"){
	writebuf(5,1);
}else if(i.cmd=="addZ"){
	writebuf(6,1);
	writebuf(i.count,2);
}else if(i.cmd=="placeBlock"){
	writebuf(7,1);
	writebuf(i.blockID,2);
	writebuf(i.blockData,2);
}else if(i.cmd=="Z++"){
	writebuf(8,1);
}else{
	writebuf(9,1);//NOP
}
jumpX 10
jumpY 11
jumpZ 12
reserved 13

*X++  14
*X--  15
*Y++  16
*Y--  17
*Z++  18
*Z--  19
*addX 20
*addBigX 21
*addY 22
*addBigY 23
*addZ 24
*addBigZ 25
assignCommandBlockData 26
placeCommandBlockWithData 27
addSmallX 28
addSmallY 29
addSmallZ 30

end 88
isSigned    90
*/

func (bdump *BDumpLegacy) formatBlocks() {
	min := []int{2147483647, 2147483647, 2147483647}
	for _, mdl := range bdump.Blocks {
		if mdl.Point.X < min[0] {
			min[0] = mdl.Point.X
		}
		if mdl.Point.Y < min[1] {
			min[1] = mdl.Point.Y
		}
		if mdl.Point.Z < min[2] {
			min[2] = mdl.Point.Z
		}
	}
	for _, mdl := range bdump.Blocks {
		mdl.Point.X -= min[0]
		mdl.Point.Y -= min[1]
		mdl.Point.Z -= min[2]
	}
}

func (bdump *BDumpLegacy) writeHeader(w *bytes.Buffer) error {
	_, err := w.Write([]byte("BDX"))
	if err != nil {
		return err
	}
	// 内部文件头
	_, err = w.Write([]byte{0})
	if err != nil {
		return err
	}
	_, err = w.Write([]byte{0})
	if err != nil {
		return err
	}
	// 写入作者之名
	// 注：现在不再写入作者信息
	_, err = w.Write([]byte{0x1f, 0x75})
	return err
	// 放置容器需要用到 117 号的 RunTimeId 调色板表
}

func (bdump *BDumpLegacy) writeBlocks(w *bytes.Buffer) error {
	bdump.formatBlocks()
	brushPosition := []int{0, 0, 0}
	blocksPalette := make(map[string]int)
	cursor := 0
	writer := &BDumpWriter{writer: w}
	for _, mdl := range bdump.Blocks {
		blknm := *mdl.Block.Name
		_, found := blocksPalette[blknm]
		if found {
			continue
		}
		err := writer.WriteCommand(&command.CreateConstantString{
			ConstantString: blknm,
		})
		if err != nil {
			return err
		}
		blocksPalette[blknm] = cursor
		cursor++
	}
	for _, mdl := range bdump.Blocks {
		for {
			if mdl.Point.X != brushPosition[0] {
				if mdl.Point.X-brushPosition[0] == 1 {
					err := writer.WriteCommand(&command.AddXValue{})
					if err != nil {
						return err
					}
				} else if mdl.Point.X-brushPosition[0] == -1 {
					err := writer.WriteCommand(&command.SubtractXValue{})
					if err != nil {
						return err
					}
				} else {
					wrap := mdl.Point.X - brushPosition[0]
					if wrap < -32768 || wrap > 32767 {
						err := writer.WriteCommand(&command.AddInt32XValue{
							Value: int32(wrap),
						})
						if err != nil {
							return err
						}
					} else if wrap < -127 || wrap > 127 {
						err := writer.WriteCommand(&command.AddInt16XValue{
							Value: int16(wrap),
						})
						if err != nil {
							return err
						}
					} else {
						err := writer.WriteCommand(&command.AddInt8XValue{
							Value: int8(wrap),
						})
						if err != nil {
							return err
						}
					}
				}
				brushPosition[0] = mdl.Point.X
				continue
			} else if mdl.Point.Y != brushPosition[1] {
				if mdl.Point.Y-brushPosition[1] == 1 {
					err := writer.WriteCommand(&command.AddYValue{})
					if err != nil {
						return err
					}
				} else if mdl.Point.Y-brushPosition[1] == -1 {
					err := writer.WriteCommand(&command.SubtractYValue{})
					if err != nil {
						return err
					}
				} else {
					wrap := mdl.Point.Y - brushPosition[1]
					if wrap > 32767 || wrap < -32768 {
						err := writer.WriteCommand(&command.AddInt32YValue{
							Value: int32(wrap),
						})
						if err != nil {
							return err
						}
					} else if wrap > 127 || wrap < -127 {
						err := writer.WriteCommand(&command.AddInt16YValue{
							Value: int16(wrap),
						})
						if err != nil {
							return err
						}
					} else {
						err := writer.WriteCommand(&command.AddInt8YValue{
							Value: int8(wrap),
						})
						if err != nil {
							return err
						}
					}
				}
				brushPosition[1] = mdl.Point.Y
				continue
			} else if mdl.Point.Z != brushPosition[2] {
				if mdl.Point.Z-brushPosition[2] == 1 {
					err := writer.WriteCommand(&command.AddZValue{})
					if err != nil {
						return err
					}
				} else if mdl.Point.Z-brushPosition[2] == 1 {
					err := writer.WriteCommand(&command.SubtractZValue{})
					if err != nil {
						return err
					}
				} else {
					wrap := mdl.Point.Z - brushPosition[2]
					if wrap > 32767 || wrap < -32768 {
						err := writer.WriteCommand(&command.AddInt32ZValue{
							Value: int32(wrap),
						})
						if err != nil {
							return err
						}
					} else if wrap > 127 || wrap < -127 {
						err := writer.WriteCommand(&command.AddInt16ZValue{
							Value: int16(wrap),
						})
						if err != nil {
							return err
						}
					} else {
						err := writer.WriteCommand(&command.AddInt8ZValue{
							Value: int8(wrap),
						})
						if err != nil {
							return err
						}
					}
				}
				brushPosition[2] = mdl.Point.Z
			}
			break
		}
		// 移动画笔
		var placeNormalBlock bool = true
		// 这个选项会决定这个方块是否以普通情况放置
		// 你可能说为什么不用 continue ，因为我希望无论如何都记录方块实体数据，而且永远都记录在方块的后面
		if mdl.ChestData != nil {
			secondMap, ok := containerIndex[*mdl.Block.Name]
			if ok {
				runTimeIdIn117, ok := secondMap[int(mdl.Block.Data)]
				if ok {
					err := writer.WriteCommand(&command.PlaceRuntimeBlockWithChestDataAndUint32RuntimeID{
						BlockRuntimeID: uint32(runTimeIdIn117),
						ChestSlots:     *mdl.ChestData,
					})
					if err != nil {
						return err
					}
					placeNormalBlock = false
				}
			}
		}
		// 容器
		if mdl.CommandBlockData != nil {
			err := writer.WriteCommand(&command.PlaceCommandBlockWithCommandBlockData{
				BlockData:        uint16(mdl.Block.Data),
				CommandBlockData: mdl.CommandBlockData,
			})
			if err != nil {
				return err
			}
			placeNormalBlock = false
		}
		// 命令方块；且优先级比箱子更高一些
		if placeNormalBlock {
			if mdl.Block.BlockStates == nil {
				err := writer.WriteCommand(&command.PlaceBlock{
					BlockConstantStringID: uint16(blocksPalette[*mdl.Block.Name]),
					BlockData:             uint16(mdl.Block.Data),
				})
				if err != nil {
					return err
				}
				// 以方块数据值(附加值)为依据放置方块
			} else {
				err := writer.WriteCommand(&command.PlaceBlockWithBlockStates{
					BlockConstantStringID: uint16(blocksPalette[*mdl.Block.Name]),
					BlockStatesString:     *mdl.Block.BlockStates,
				})
				if err != nil {
					return err
				}
				// 以方块状态为依据放置方块
				// 我更推荐使用这一个方式来放置方块，因为方块数据值(附加值)已不再在新版本 MC 中使用
				// ——Happy2018new
			}
		}
		// 常规方块写入
		if mdl.NBTData != nil {
			err := writer.WriteCommand(&command.AssignNBTData{
				Data: mdl.NBTData,
			})
			if err != nil {
				return err
			}
		}
		// 写入方块实体数据
	}
	return nil
}

func (bdump *BDumpLegacy) WriteToFile(path string, localCert string, localKey string) (error, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("Failed to open file: %v", err), nil
	}
	defer file.Close()
	_, err = file.Write([]byte("BD@"))
	if err != nil {
		return fmt.Errorf("Failed to write BRBDP file header"), nil
	}
	buffer := &bytes.Buffer{}
	brw := brotli.NewWriter(file)
	err = bdump.writeHeader(buffer)
	if err != nil {
		return err, nil
	}
	err = bdump.writeBlocks(buffer)
	if err != nil {
		return err, nil
	}
	bts := buffer.Bytes()
	_, err = brw.Write(bts)
	if err != nil {
		return err, nil
	}
	sign, signerr := SignBDX(bts, localKey, localCert)
	if signerr != nil {
		brw.Write([]byte("XE"))
	} else {
		brw.Write(append([]byte{88}, sign...))
		if len(sign) >= 255 {
			realLength := make([]byte, 2)
			binary.BigEndian.PutUint16(realLength, uint16(len(sign)))
			brw.Write(realLength)
			brw.Write([]byte{uint8(255)})
		} else {
			brw.Write([]byte{uint8(len(sign))})
		}
		brw.Write([]byte{90})
	}
	err = brw.Close()
	return err, signerr
}
