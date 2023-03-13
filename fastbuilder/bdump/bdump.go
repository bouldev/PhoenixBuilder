package bdump

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"phoenixbuilder/fastbuilder/bdump/command"
	"phoenixbuilder/fastbuilder/types"

	"github.com/andybalholm/brotli"
)

type BDump struct {
	Author string // Should be empty
	Blocks []*types.Module
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

func (bdump *BDump) formatBlocks() {
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

func (bdump *BDump) writeHeader(w io.Writer) error {
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
	return err
}

func (bdump *BDump) writeBlocks(w io.Writer) error {
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
		blkst := mdl.Block.BlockStates
		if len(blkst) == 0 {
			continue
		}
		_, found := blocksPalette[blkst]
		if found {
			continue
		}
		err := writer.WriteCommand(&command.CreateConstantString{
			ConstantString: blkst,
		})
		if err != nil {
			return err
		}
		blocksPalette[blkst] = cursor
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
		if mdl.ChestData != nil {
			err := writer.WriteCommand(&command.PlaceBlockWithChestData{
				BlockConstantStringID: uint16(blocksPalette[*mdl.Block.Name]),
				BlockData:             uint16(mdl.Block.Data),
				ChestSlots:            *mdl.ChestData,
			})
			if err != nil {
				return err
			}
		} else if mdl.CommandBlockData != nil {
			err := writer.WriteCommand(&command.PlaceCommandBlockWithCommandBlockData{
				BlockData:        uint16(mdl.Block.Data),
				CommandBlockData: mdl.CommandBlockData,
			})
			if err != nil {
				return err
			}
		} else if mdl.NBTData == nil {
			if len(mdl.Block.BlockStates) == 0 {
				err := writer.WriteCommand(&command.PlaceBlock{
					BlockConstantStringID: uint16(blocksPalette[*mdl.Block.Name]),
					BlockData:             uint16(mdl.Block.Data),
				})
				if err != nil {
					return err
				}
			} else {
				err := writer.WriteCommand(&command.PlaceBlockWithBlockStates{
					BlockConstantStringID:       uint16(blocksPalette[*mdl.Block.Name]),
					BlockStatesConstantStringID: uint16(blocksPalette[mdl.Block.BlockStates]),
				})
				if err != nil {
					return err
				}
			}
		} else {
			err := writer.WriteCommand(&command.PlaceBlockWithNBTData{
				BlockConstantStringID:       uint16(blocksPalette[*mdl.Block.Name]),
				BlockStatesConstantStringID: uint16(blocksPalette[mdl.Block.BlockStates]),
				BlockNBT_bytes:              mdl.NBTData,
			})
			if err != nil {
				return err
			}
		}
		/*
			if mdl.NBTData != nil {
				err := writer.WriteCommand(&command.AssignNBTData{
					Data: mdl.NBTData,
				})
				if err != nil {
					return err
				}
			}
		*/
	}
	return nil
}

func (bdump *BDump) WriteToFile(path string, localCert string, localKey string) (error, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("Failed to open file: %v", err), nil
	}
	defer file.Close()
	_, err = file.Write([]byte("BD@"))
	if err != nil {
		return fmt.Errorf("Failed to write BRBDP file header"), nil
	}
	brw := brotli.NewWriter(file)
	brhw := &HashedWriter{
		writer: brw,
		hash:   sha256.New(),
	}
	err = bdump.writeHeader(brhw)
	if err != nil {
		return err, nil
	}
	err = bdump.writeBlocks(brhw)
	if err != nil {
		return err, nil
	}
	fileHash := brhw.hash.Sum(nil)
	sign, signerr := SignBDX(fileHash, localKey, localCert)
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
