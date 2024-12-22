package builder

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
	"compress/gzip"
	"fmt"
	"io"
	"os"
	I18n "phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"

	"github.com/Tnze/go-mc/nbt"
)

/*
SplitResult结构体表示一个分割后的结果
Chunks:        存储了分割后的"块"模块
Horizontal:    表示整个建筑水平方向上的"块"模块数量
Vertical:      表示整个建筑垂直方向上的"块"模块数量
*/
type SplitResult struct {
	Chunks     []ChunkModule
	Horizontal int
	Vertical   int
}

/*
ChunkModule结构体表示一个分割后的"块"模块
Blocks:        存储了"块"模块内的方块数据
Data:          存储了"块"模块内的方块数据(特殊值?)
SizeX:         由于移除了扩充建筑大小的操作，导致我们不能确定分割后的"块"平面大小绝对为16x16，所以我们需要记录下每个"块"模块的实际大小，SizeX为该"块"的实际宽度(X轴)
SizeZ:   	   由于移除了扩充建筑大小的操作，导致我们不能确定分割后的"块"平面大小绝对为16x16，所以我们需要记录下每个"块"模块的实际大小，SizeZ为该"块"的实际宽度(Z轴)
Position:      表示"块"模块的起始位置，用于计算Position与Position_End两点所包括的所有方块
Position_End:  表示"块"模块的终止位置，用于计算Position与Position_End两点所包括的所有方块
*/
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

/*
SplitChunks函数用于将整个建筑分割成多个"块"模块

width:      表示建筑的宽度
height:     表示建筑的高度
length:     表示建筑的长度
X:          表示导入点的起始X坐标
Y:          表示导入点的起始Y坐标
Z:          表示导入点的起始Z坐标
schBlocks:  表示整个建筑内的方块数据
schData:    表示整个建筑内的方块数据(特殊值?)

该函数无法较好地使用概括的方式进行描述，具体的分割逻辑需要根据分行注释仔细阅读函数内部的代码
*/
func SplitChunks(width, height, length, X, Y, Z int, schBlocks, schData []byte) SplitResult {
	var chunks []ChunkModule
	// 定义一个匿名函数getMin，用于获取两个整数中的较小值，该函数用于确保在分割时不会超出数组的边界
	getMin := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	// 该循环以Z轴方向对整个建筑进行切块处理，每次切块的步长为16，与Minecraft中一个区块(16x16)的边长一致
	for z := 0; z < length; z += 16 {
		// 通过getMin函数确定当前切块在Z轴上的的实际结束位置，在后面用于防止超出整个区域在Z轴方向的长度(length)边界
		endZ := getMin(z+16, length)
		// 该循环以X轴方向对整个建筑进行切块处理，每次切块的步长为16，与Minecraft中一个区块(16x16)的边长一致
		// Y轴方向显然是不必要的，它的大小与建筑的高度(即height)一致
		for x := 0; x < width; x += 16 {
			// 通过getMin函数确定当前切块在X轴上的的实际结束位置，在后面用于防止超出整个区域在X轴方向的长度(width)边界
			endX := getMin(x+16, width)
			chunk := ChunkModule{
				SizeX:    endX - x,
				SizeZ:    endZ - z,
				Position: types.Position{X: x, Z: z},
				// 如果当前"块"的结束位置(endX，endZ)小于整个区域在对应轴方向的长度(width，length)，则当前"块"的endX，endZ就为当前"块"先前通过步长16计算得到的endX和endZ
				// 相反的，如果当前"块"的结束位置(endX，endZ)大于整个区域在对应轴方向的长度(width，length)，说明建筑平面内的横，纵方向上无法以步长16进行切割，因此当前"块"的endX，endZ就为整个区域在对应轴方向的长度(width，length)
				// 最后，由于我们将坐标轴上的每个点看作Minecraft世界内的一个方块，因此我们需要将当前"块"的endX，endZ减去1，以确保当前"块"的endX，endZ是当前"块"的最后一个方块的坐标
				Position_End: types.Position{X: endX - 1, Y: height, Z: endZ - 1},
			}
			//用于判定是否是"块"内第一个方块
			PrevBlockIndex := 0
			// 嵌套的多层循环用于遍历当前区块模块内部的每一个方块的位置(在X、Y、Z三轴上)
			for y2 := chunk.Position.Y; y2 <= chunk.Position_End.Y; y2++ {
				for z2 := chunk.Position.Z; z2 <= chunk.Position_End.Z; z2++ {
					for x2 := chunk.Position.X; x2 <= chunk.Position_End.X; x2++ {
						var BlockIndex int
						// 如果是"块"内第一个方块，那么该"块"中第一个方块的BlockIndex就为x2 + z2*width + y2*width*length，否则就为前一个方块的BlockIndex+1
						// 这样一来我们就无需为每个方块单独计算该方块的BlockIndex，而是可以通过前一个方块的BlockIndex+1来得到当前方块的BlockIndex，节省了大量的计算时间
						if PrevBlockIndex == 0 {
							BlockIndex = x2 + z2*width + y2*width*length
						} else {
							BlockIndex = PrevBlockIndex + 1
							PrevBlockIndex = BlockIndex
						}
						// 判定当前方块的BlockIndex是否在整个建筑的范围内，如果在范围内，则将当前方块的方块数据(schBlocks[BlockIndex])和方块数据(schData[BlockIndex])添加到当前"块"模块的Blocks和Data切片中
						if BlockIndex >= 0 && BlockIndex < len(schBlocks) {
							chunk.Blocks = append(chunk.Blocks, schBlocks[BlockIndex])
							chunk.Data = append(chunk.Data, schData[BlockIndex])
						}
					}
				}
			}
			// 提前加上导入点的坐标，在之后的使用中，我们可以直接获取当前"块"模块相对于整个世界的坐标，而无需再次计算
			chunk.Position = types.Position{X: X + x, Y: Y, Z: Z + z}
			chunk.Position_End = types.Position{X: X + x + 15, Y: height + Y, Z: Z + z + 15}
			chunks = append(chunks, chunk)
		}
	}
	// 计算横向和纵向的"块"数量
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
getChunkSequence函数用于正确地排列分割后的"块"模块的导入顺序，以避免在导入时出现区块未加载从而导致导入不全的问题

totalElements:       表示总区块数量
horizontalElements:  表示横向区块数
verticalElements:    表示纵向区块数

根据提供的totalElements(区块总量)、horizontalElements(横向区块数)和verticalElements(纵向区块数)，返回一个以特定顺序排列的整数切片
顺序的规则为: 对于每一行，如果是奇数行，则从右到左排列区块索引，否则从左到右排列区块索引

例如，对于totalElements=12、horizontalElements=4和verticalElements=3，我们可以将其原始排列抽象为一个二维数组:
{0, 1, 2, 3},
{4, 5, 6, 7},
{8, 9, 10, 11}
即总共12个区块，每行有4个区块，共有3行区块。
对于第0行(偶数行)，我们从左到右排列，得到{0, 1, 2, 3}
对于第1行(奇数行)，我们从右到左排列，得到{7, 6, 5, 4}
对于第2行(偶数行)，我们从左到右排列，得到{8, 9, 10, 11}
最后合并得到最终的结果{0, 1, 2, 3, 7, 6, 5, 4, 8, 9, 10, 11}，这就是getChunkSequence函数的返回值
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
