package builder

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"

	"github.com/Tnze/go-mc/nbt"
)

type ChunkModule struct {
	Blocks       []byte
	Data         []byte
	Position     types.Position
	Position_End types.Position
}

func Schematic(config *types.MainConfig, blc chan *types.Module) error {
	file, err := os.Open(config.Path)
	if err != nil {
		return I18n.ProcessSystemFileError(err)
	}
	defer file.Close()
	gzip, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzip.Close()
	buffer, err := ioutil.ReadAll(gzip)

	var SchematicModule struct {
		Blocks    []byte `nbt:"Blocks"`
		Data      []byte `nbt:"Data"`
		Width     int    `nbt:"Width"`
		Length    int    `nbt:"Length"`
		Height    int    `nbt:"Height"`
		WEOffsetX int    `nbt:"WEOffsetX"`
		WEOffsetY int    `nbt:"WEOffsetY"`
		WEOffsetZ int    `nbt:"WEOffsetZ"`
	}

	if err := nbt.Unmarshal(buffer, &SchematicModule); err != nil {
		// Won't return the error `err` since it contains a large content that can
		// crash the server after being sent.
		return fmt.Errorf(I18n.T(I18n.Sch_FailedToResolve))
	}
	if len(SchematicModule.Blocks) == 0 {
		return fmt.Errorf("Invalid structure.")
	}
	Size := [3]int{SchematicModule.Width, SchematicModule.Height, SchematicModule.Length}
	var FixedSchematicModule struct {
		Blocks []byte
		Data   []byte
		Width  int
		Height int
		Length int
	}
	FixedSchematicModule.Width = roundUpToNearestMultipleOf16(Size[0])
	FixedSchematicModule.Height = Size[1]
	FixedSchematicModule.Length = roundUpToNearestMultipleOf16(Size[2])
	FixedSize := [3]int{FixedSchematicModule.Width, FixedSchematicModule.Height, FixedSchematicModule.Length}
	FixedSchematicModule.Blocks = expandAndFillWithAir(SchematicModule.Blocks, Size, FixedSize)
	FixedSchematicModule.Data = expandAndFillWithAir(SchematicModule.Data, Size, FixedSize)
	chunks := detachChunks(FixedSize[0], FixedSize[1], FixedSize[2], config.Position.X, config.Position.Y, config.Position.Z, FixedSchematicModule.Blocks, FixedSchematicModule.Data)
	BlockIndex := 0
	for id := 0; id < len(chunks); id++ {
		BlockIndex = 0
		for y := 0; y < FixedSize[1]; y++ {
			for z := 0; z < 16; z++ {
				for x := 0; x < 16; x++ {
					p := chunks[id].Position
					p.X += x
					p.Y += y
					p.Z += z
					var b types.Block
					b.Name = &BlockStr[chunks[id].Blocks[BlockIndex]]
					b.Data = uint16(chunks[id].Data[BlockIndex])
					if *b.Name != "air" {
						blc <- &types.Module{Point: p, Block: &b}
					}
					BlockIndex++
				}
			}
		}
	}
	return nil
}

func roundUpToNearestMultipleOf16(num int) int { //向上舍入到最接近的 16 的倍数
	remainder := num % 16
	if remainder == 0 {
		return num
	}
	return num + 16 - remainder
}

func expandAndFillWithAir(blocks []byte, original [3]int, new [3]int) []byte { //修补文件并填充空气
	newBlocks := make([]byte, new[0]*new[1]*new[2])
	for y := 0; y < new[1]; y++ {
		for z := 0; z < new[2]; z++ {
			for x := 0; x < new[0]; x++ {
				if x < original[0] && y < original[1] && z < original[2] {
					newBlocks[y*new[0]*new[2]+z*new[0]+x] = blocks[y*original[0]*original[2]+z*original[0]+x]
				} else {
					newBlocks[y*new[0]*new[2]+z*new[0]+x] = 0
				}
			}
		}
	}
	return newBlocks
}

func detachChunks(width, height, length, X, Y, Z int, schBlocks, schData []byte) []ChunkModule { //拆分区块 这个地方不知道中了什么邪，改参数名就炸
	size := 16
	var chunks []ChunkModule
	for z := 0; z < length; z += size {
		for x := 0; x < width; x += size {
			chunk := ChunkModule{
				Position:     types.Position{X: x, Y: 0, Z: z},
				Position_End: types.Position{X: x + size - 1, Y: height, Z: z + size - 1},
			}
			for y2 := chunk.Position.Y; y2 <= chunk.Position_End.Y; y2++ {
				for z2 := chunk.Position.Z; z2 <= chunk.Position_End.Z; z2++ {
					for x2 := chunk.Position.X; x2 <= chunk.Position_End.X; x2++ {
						BlockIndex := x2 + z2*width + y2*width*length
						if BlockIndex >= 0 && BlockIndex < len(schBlocks) {
							chunk.Blocks = append(chunk.Blocks, schBlocks[BlockIndex])
							chunk.Data = append(chunk.Data, schData[BlockIndex])
						}
					}
				}
			}
			chunk.Position = types.Position{X: X + x, Y: Y, Z: Z + z}
			chunk.Position_End = types.Position{X: X + x + size - 1, Y: height + Y, Z: Z + z + size - 1}
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}
