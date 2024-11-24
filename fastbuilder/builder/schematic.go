package builder

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"

	"github.com/Tnze/go-mc/nbt"
)

type SplitResult struct {
	Chunks     []ChunkModule
	Horizontal int
	Vertical   int
}

type ChunkModule struct {
	Blocks       []byte
	Data         []byte
	SizeX        int
	SizeZ        int
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
	buffer, err := io.ReadAll(gzip)

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
	Offset := [3]int{SchematicModule.WEOffsetX, SchematicModule.WEOffsetY, SchematicModule.WEOffsetZ}
	split := SplitChunks(Size[0], Size[1], Size[2], config.Position.X, config.Position.Y, config.Position.Z, SchematicModule.Blocks, SchematicModule.Data)
	sequence := getChunkSequence(len(split.Chunks), split.Horizontal, split.Vertical)
	BlockIndex := 0
	for id := 0; id < len(split.Chunks); id++ {
		BlockIndex = 0
		for y := 0; y < Size[1]; y++ {
			for z := 0; z < split.Chunks[sequence[id]].SizeZ; z++ {
				for x := 0; x < split.Chunks[sequence[id]].SizeX; x++ {
					p := split.Chunks[sequence[id]].Position
					p.X += x + Offset[0]
					p.Y += y + Offset[1]
					p.Z += z + Offset[2]
					var b types.Block
					b.Name = &BlockStr[split.Chunks[sequence[id]].Blocks[BlockIndex]]
					b.Data = uint16(split.Chunks[sequence[id]].Data[BlockIndex])
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

func SplitChunks(width, height, length, X, Y, Z int, schBlocks, schData []byte) SplitResult {
	var chunks []ChunkModule
	getMin := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	for z := 0; z < length; z += 16 {
		endZ := getMin(z+16, length)
		for x := 0; x < width; x += 16 {
			endX := getMin(x+16, width)
			chunk := ChunkModule{
				SizeX:        endX - x,
				SizeZ:        endZ - z,
				Position:     types.Position{X: x, Z: z},
				Position_End: types.Position{X: endX - 1, Y: height, Z: endZ - 1},
			}
			PrevBlockIndex := 0
			for y2 := chunk.Position.Y; y2 <= chunk.Position_End.Y; y2++ {
				for z2 := chunk.Position.Z; z2 <= chunk.Position_End.Z; z2++ {
					for x2 := chunk.Position.X; x2 <= chunk.Position_End.X; x2++ {
						var BlockIndex int
						if PrevBlockIndex == 0 {
							BlockIndex = x2 + z2*width + y2*width*length
						} else {
							BlockIndex = PrevBlockIndex + 1
							PrevBlockIndex = BlockIndex
						}
						if BlockIndex >= 0 && BlockIndex < len(schBlocks) {
							chunk.Blocks = append(chunk.Blocks, schBlocks[BlockIndex])
							chunk.Data = append(chunk.Data, schData[BlockIndex])
						}
					}
				}
			}
			chunk.Position = types.Position{X: X + x, Y: Y, Z: Z + z}
			chunk.Position_End = types.Position{X: X + x + 15, Y: height + Y, Z: Z + z + 15}
			chunks = append(chunks, chunk)
		}
	}
	horizontalCount := width / 16
	if width%16 != 0 {
		horizontalCount++
	}
	verticalCount := length / 16
	if length%16 != 0 {
		verticalCount++
	}
	return SplitResult{
		Chunks:     chunks,
		Horizontal: horizontalCount,
		Vertical:   verticalCount,
	}
}

/*
获取正确导入顺序 totalElements: 分割块总数 horizontalElements: 水平分割块数 verticalElements: 垂直分割块数
假设一个建筑的区块排列是
0  1  2  3  4
5  6  7  8  9
那么要按照"之"字形导入,则区块顺序应该是
0  1  2  3  4  9  8  7  6  5
以下代码实现了重新排列顺序
*/
func getChunkSequence(totalElements, horizontalElements, verticalElements int) []int {
	result := make([]int, totalElements)
	index := 0
	for i := 0; i < verticalElements; i++ {
		start := i * horizontalElements
		end := start + horizontalElements
		if i%2 == 0 {
			for j := start; j < end; j++ {
				result[index] = j
				index++
			}
		} else {
			for j := end - 1; j >= start; j-- {
				result[index] = j
				index++
			}
		}
	}
	return result
}
