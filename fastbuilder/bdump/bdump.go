package bdump

/*
 * This file is part of PhoenixBuilder.

 * PhoenixBuilder is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.

 * Copyright (C) 2021-2025 Bouldev
 */

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"phoenixbuilder/fastbuilder/bdump/command"
	"phoenixbuilder/fastbuilder/types"

	"github.com/andybalholm/brotli"
)

type BDump struct {
	Author string // Deprecated field, would generally be empty
	Blocks []*types.Module
}

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
	}
	return nil
}

func (bdump *BDump) WriteToFile(path string) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("WriteToFile: Failed to open file: %v", err)
	}
	defer file.Close()
	_, err = file.Write([]byte("BD@"))
	if err != nil {
		return fmt.Errorf("WriteToFile: Failed to write BRBDP file header")
	}
	brw := brotli.NewWriter(file)
	brhw := &HashedWriter{
		writer: brw,
		hash:   sha256.New(),
	}
	err = bdump.writeHeader(brhw)
	if err != nil {
		return err
	}
	err = bdump.writeBlocks(brhw)
	if err != nil {
		return err
	}
	brw.Write([]byte("XE"))
	err = brw.Close()
	return err
}
