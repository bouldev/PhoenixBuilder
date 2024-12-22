package mcstructure

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
	"fmt"
	"math"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/minecraft/nbt"
	"strconv"
	"strings"
)

type Area struct {
	BeginX int32
	BeginY int32
	BeginZ int32
	SizeX  int32
	SizeY  int32
	SizeZ  int32
}

type AreaLocation [2]int

type BlockPos [3]int32

type Mcstructure struct {
	area                     Area
	blockPalette             []string
	blockPalette_blockStates []string
	// ^ NOTE: All TAG_BYTE values are treated as booleans
	blockPalette_blockData []int16
	foreground             []int16
	background             []int16
	blockNBT               map[int]map[string]interface{}
}

/*
用于拆分一个大区域为若干个小区域；当 useSpecialSplitWay 为真时，将蛇形拆分区域

返回值 []Area 代表一个已经排好顺序的若干个小区域。
返回值 map[AreaLocation]int 代表可以通过 区域坐标(AreaLocation) 来访问 []Area 的对应项。
返回值 map[int]AreaLocation 是返回值 map[AreaLocation]int 的逆过程
*/
func SplitArea(beginPos BlockPos, endPos BlockPos, splitSizeX int32, splitSizeZ int32, useSpecialSplitWay bool) ([]Area, map[AreaLocation]int, map[int]AreaLocation) {
	if splitSizeX < 0 {
		splitSizeX = -splitSizeX
	}
	if splitSizeZ < 0 {
		splitSizeZ = -splitSizeZ
	}
	// 考虑一些特殊的情况，此举是为了更高的兼容性
	if endPos[0] < beginPos[0] {
		tmp := beginPos[0]
		beginPos[0] = endPos[0]
		endPos[0] = tmp
	}
	if endPos[1] < beginPos[1] {
		tmp := beginPos[1]
		beginPos[1] = endPos[1]
		endPos[1] = tmp
	}
	if endPos[2] < beginPos[2] {
		tmp := beginPos[2]
		beginPos[2] = endPos[2]
		endPos[2] = tmp
	}
	// 考虑一些特殊的情况，此举是为了更高的兼容性
	sizeX := endPos[0] - beginPos[0] + 1
	sizeY := endPos[1] - beginPos[1] + 1
	sizeZ := endPos[2] - beginPos[2] + 1
	// 取得 Area 的大小
	chunkX_length := int(math.Ceil(float64(sizeX) / float64(splitSizeX)))
	chunkZ_length := int(math.Ceil(float64(sizeZ) / float64(splitSizeZ)))
	// 取得各轴上需要拆分的区域数
	ret := make([]Area, chunkX_length*chunkZ_length) // 这个东西最终会 return 掉
	areaLoctionToInt := map[AreaLocation]int{}       // 知道了区域的坐标求区域在 []Area 的位置
	IntToareaLoction := map[int]AreaLocation{}       // 知道了区域在 []Area 的位置求区域坐标
	facing := -1                                     // 蛇形处理的时候需要用到这个
	key := -1                                        // 向 ans 插入数据的时候需要用到这个
	for chunkX := 1; chunkX <= chunkX_length; chunkX++ {
		facing = facing * -1
		BeginX := splitSizeX*(int32(chunkX)-1) + beginPos[0]
		xLength := splitSizeX
		if BeginX+xLength-1 > endPos[0] {
			xLength = endPos[0] - BeginX + 1
		}
		for chunkZ := 1; chunkZ <= chunkZ_length; chunkZ++ {
			key++
			currentChunkZ := chunkZ
			if useSpecialSplitWay && facing == -1 {
				currentChunkZ = chunkZ_length - currentChunkZ + 1
			}
			BeginZ := splitSizeZ*(int32(currentChunkZ)-1) + beginPos[2]
			zLength := splitSizeZ
			if BeginZ+zLength-1 > endPos[2] {
				zLength = endPos[2] - BeginZ + 1
			}
			ret[key] = Area{
				BeginX: BeginX,
				BeginY: beginPos[1],
				BeginZ: BeginZ,
				SizeX:  xLength,
				SizeY:  sizeY,
				SizeZ:  zLength,
			}
			areaLoctionToInt[AreaLocation{chunkX - 1, currentChunkZ - 1}] = key
			IntToareaLoction[key] = AreaLocation{chunkX - 1, currentChunkZ - 1}
		}
	}
	return ret, areaLoctionToInt, IntToareaLoction
}

func GetMCStructureData(area Area, structure map[string]interface{}) (Mcstructure, error) {
	blockPalette := []string{}
	blockPalette_blockStates := []string{}
	blockPalette_blockData := []int16{}
	blockNBT := map[int]map[string]interface{}{}
	foreground := []int16{}
	background := []int16{}
	_, ok := structure["structure"]
	if !ok {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"]; structure = %#v", structure)
	}
	value_structure, normal := structure["structure"].(map[string]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"]; structure = %#v", structure)
	}
	// structure["structure"]
	_, ok = value_structure["palette"]
	if !ok {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"]; structure = %#v", value_structure)
	}
	value_palette, normal := value_structure["palette"].(map[string]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"]; structure = %#v", value_structure)
	}
	// structure["structure"]["palette"]
	_, ok = value_palette["default"]
	if !ok {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"]; palette = %#v", value_palette)
	}
	value_default, normal := value_palette["default"].(map[string]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"]; palette = %#v", value_palette)
	}
	// structure["structure"]["palette"]["default"]
	_, ok = value_default["block_palette"]
	if !ok {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"]; default = %#v", value_default)
	}
	value_block_palette, normal := value_default["block_palette"].([]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"]; default = %#v", value_default)
	}
	// structure["structure"]["palette"]["default"]["block_palette"]
	for key, value := range value_block_palette {
		got, normal := value.(map[string]interface{})
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"name\"]; block_palette[%v] = %#v", key, key, value_block_palette[key])
		}
		// 这里确认下数据类型，就是这个 got 必须得是个复合标签
		_, ok = got["name"]
		if !ok {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"name\"]; block_palette[%v] = %#v", key, key, value_block_palette[key])
		}
		value_name, normal := got["name"].(string)
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"name\"]; block_palette[%v] = %#v", key, key, value_block_palette[key])
		}
		blockPalette = append(blockPalette, value_name)
		// 得到方块的名称
		// 这里的名称是携带了命名空间 minecraft 的
		// 命名空间会在后边删掉
		_, ok = got["states"]
		if !ok {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"states\"]; block_palette[%v] = %#v", key, key, value_block_palette[key])
		}
		value_states, normal := got["states"].(map[string]interface{})
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"states\"]; block_palette[%v] = %#v", key, key, value_block_palette[key])
		}
		blockStates, err := MarshalBlockStates(value_states)
		if err != nil {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"states\"]; block_palette[%v] = %#v", key, key, value_block_palette[key])
		}
		blockPalette_blockStates = append(blockPalette_blockStates, blockStates)
		// 得到方块的方块状态
		_, ok = got["val"]
		if !ok {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"val\"]; block_palette[%v] = %#v", key, key, value_block_palette[key])
		}
		val, normal := got["val"].(int16)
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"val\"]; block_palette[%v] = %#v", key, key, value_block_palette[key])
		}
		blockPalette_blockData = append(blockPalette_blockData, val)
		// 得到方块的方块数据值(附加值)
	}
	// 这个 for 用于获取调色板的信息的
	_, ok = value_default["block_position_data"]
	if ok {
		value_block_position_data, normal := value_default["block_position_data"].(map[string]interface{})
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"default\"][\"block_position_data\"]; default = %#v", value_default)
		}
		for key, value := range value_block_position_data {
			block_position_data, ok := value.(map[string]interface{})
			if !ok {
				return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"default\"][\"block_position_data\"][%v]; block_position_data[%v] = %#v", key, key, block_position_data[key])
			}
			location_of_block_position_data, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"default\"][\"block_position_data\"][%v]; block_position_data[%v] = %#v", key, key, block_position_data[key])
			}
			if blockNBT[int(location_of_block_position_data)] == nil {
				blockNBT[int(location_of_block_position_data)] = make(map[string]interface{})
			}
			blockNBT[int(location_of_block_position_data)] = map[string]interface{}{"block_position_data": block_position_data}
		}
	}
	// 众所不一定周知，这个方块实体数据可能是不存在的(当然这个我没测试过)
	// 然后找到所有的方块实体数据，放于 map(blockNBT) 中
	_, ok = value_structure["block_indices"]
	if !ok {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"block_indices\"]; structure = %#v", structure)
	}
	value_block_indices, normal := value_structure["block_indices"].([]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"block_indices\"]; structure = %#v", structure)
	}
	// structure["structure"]["block_indices"]
	if len(value_block_indices) != 2 {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"block_indices\"]; structure = %#v", structure)
	}
	// 这里要求 structure["structure"]["block_indices"] 的长度必须为 2
	// 毕竟是由 前景层 和 背景层 的索引所制成的两张表
	value_block_indices_0, normal := value_block_indices[0].([]int32)
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"block_indices\"][0]; block_indices = %#v", value_block_indices)
	}
	for _, blockLocation := range value_block_indices_0 {
		foreground = append(foreground, int16(blockLocation))
	}
	// 这里先拿前景层方块的索引表
	value_block_indices_1, normal := value_block_indices[1].([]int32)
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Failed on structure[\"structure\"][\"block_indices\"][1]; block_indices = %#v", value_block_indices)
	}
	for _, blockLocation := range value_block_indices_1 {
		background = append(background, int16(blockLocation))
	}
	// 然后再去拿背景层方块的索引表
	return Mcstructure{
		area:                     area,
		blockPalette:             blockPalette,
		blockPalette_blockStates: blockPalette_blockStates,
		blockPalette_blockData:   blockPalette_blockData,
		foreground:               foreground,
		background:               background,
		blockNBT:                 blockNBT,
	}, nil
}

// 根据 mcstructure 的起点和尺寸，以及提供的方块坐标，
// 寻找这个方块在 mcstructure 中的角标
func SearchForBlock(structureInfo Area, pos BlockPos) (int, error) {
	pos[0] -= structureInfo.BeginX
	pos[1] -= structureInfo.BeginY
	pos[2] -= structureInfo.BeginZ
	// 将方块的绝对坐标转换为相对坐标(相对于 mcstructure)
	blockCount := structureInfo.SizeX * structureInfo.SizeY * structureInfo.SizeZ
	// 计算结构的尺寸
	angleMark := structureInfo.SizeY*structureInfo.SizeZ*pos[0] + structureInfo.SizeZ*pos[1] + pos[2]
	// 计算方块相对于 mcstructure 的角标
	if angleMark > blockCount-1 {
		return -1, fmt.Errorf("SearchForBlock: Index out of the list, occurred in input[%v]", angleMark)
	}
	return int(angleMark), nil
}

/*
基于区块的大小对整个待导出区域进行重排，并写入对应的方块、NBT数据。

我们对整个待导出区域按 64*64 大小拆分，且蛇形拆分(使用SplitArea拆分)，
然后再获取拆分得到的各个小区域的 mcstructure 数据，
然后处理后制成此函数中的 allChunks 表。

chunkPosIndicator 用于通过 区域坐标 来查这个区域在 allChunks 表的位置。

currentExport 指代当前 Task 指定的导出区域，
也就是根据 set(get) 和 setend(get end) 制成的 Area
*/
func DumpBlocks(
	allAreas []Mcstructure,
	reversedMap map[AreaLocation]int,
	currentExport Area,
) ([]*types.Module, error) {
	ans := make([]*types.Module, 0)
	// ans 最终将会被 return
	allChunks, _, chunkPosIndicator := SplitArea(
		BlockPos{currentExport.BeginX, currentExport.BeginY, currentExport.BeginZ},
		BlockPos{
			currentExport.BeginX + currentExport.SizeX - 1,
			currentExport.BeginY + currentExport.SizeY - 1,
			currentExport.BeginZ + currentExport.SizeZ - 1,
		},
		16, 16, true,
	)
	// 将所有待导出区域按 16*16 的大小拆分为区块，且蛇形拆分
	// 然后按照得到的结果重排处理
	for key, value := range allChunks {
		chunkPos := chunkPosIndicator[key]
		chunkPos[0] = int(math.Floor(float64(chunkPos[0]) / 4))
		chunkPos[1] = int(math.Floor(float64(chunkPos[1]) / 4))
		// 取得当前遍历的区块的坐标
		// 这里已经把坐标变换到 allAreas 下的坐标系中
		targetAreaPos := reversedMap[chunkPos]
		targetArea := allAreas[targetAreaPos]
		// 取得被遍历区块对应的 mcstructure
		i, _, _ := SplitArea(
			BlockPos{value.BeginX, value.BeginY, value.BeginZ},
			BlockPos{
				value.BeginX + value.SizeX - 1,
				value.BeginY + value.SizeY - 1,
				value.BeginZ + value.SizeZ - 1,
			},
			1, 1, true,
		)
		allBlocksInCurrentChunk := make([]int32, 0)
		for _, val := range i {
			got, err := SearchForBlock(targetArea.area, BlockPos{
				val.BeginX,
				val.BeginY,
				val.BeginZ,
			})
			if err != nil {
				return []*types.Module{}, fmt.Errorf("DumpBlocks: %v", err)
			}
			allBlocksInCurrentChunk = append(allBlocksInCurrentChunk, int32(got))
		}
		// 枚举出被遍历区块中所有方块的坐标(只枚举其中一层)
		for key, val := range allBlocksInCurrentChunk {
			val -= int32(targetArea.area.SizeZ)
			// 这个前置处理方法可能不太优雅
			// 凑合着用吧
			for j := int32(0); j < targetArea.area.SizeY; j++ {
				val += int32(targetArea.area.SizeZ)
				// 前往下一层
				foreground_blockName := "undefined"
				background_blockName := "undefined"
				foreground_blockStates := "undefined"
				background_blockStates := "undefined"
				// 初始化
				fgId := targetArea.foreground[val] // 前景层方块在调色板中的id
				bgId := targetArea.background[val] // 背景层方块在调色板中的id
				if fgId != -1 {
					foreground_blockName = strings.Replace(targetArea.blockPalette[fgId], "minecraft:", "", 1) // 前景层方块的名称
					foreground_blockStates = targetArea.blockPalette_blockStates[fgId]                         // 前景层方块的方块状态
				}
				if bgId != -1 {
					background_blockName = strings.Replace(targetArea.blockPalette[bgId], "minecraft:", "", 1) // 背景层方块的名称
					background_blockStates = targetArea.blockPalette_blockStates[bgId]                         // 背景层方块的方块状态
				}
				if fgId == -1 && bgId == -1 {
					foreground_blockName = "structure_void"
					foreground_blockStates = "[]"
				}
				// 获得基本信息
				var hasNBT bool = false
				var blockNBT []byte
				var err error = nil
				// 变量初始化
				// 危险！变量初始化这里不要动，不然可能会出现一些意想不到的 Bug
				got, ok := targetArea.blockNBT[int(val)]
				if ok {
					_, ok := got["block_position_data"]
					if !ok {
						return []*types.Module{}, fmt.Errorf("DumpBlocks: Crashed by could not found \"block_position_data\", occurred in %#v", targetArea.blockNBT[int(val)])
					}
					block_position_data, normal := got["block_position_data"].(map[string]interface{})
					if !normal {
						return []*types.Module{}, fmt.Errorf("DumpBlocks: Crashed by invalid \"block_position_data\", occurred in %#v", got["block_position_data"])
					}
					// 只要这个方块被记录了 NBT 数据，那么一定会有 block_position_data
					_, ok = block_position_data["block_entity_data"]
					// 虽然这个方块被记录了 NBT 数据，但不一定是一个方块实体
					// 所以即便这个被记录了 NBT 的方块没有 block_entity_data ，也不应该返回错误
					if ok {
						block_entity_data, normal := block_position_data["block_entity_data"].(map[string]interface{})
						if !normal {
							return []*types.Module{}, fmt.Errorf("DumpBlocks: Crashed by invalid \"block_entity_data\", occurred in %#v", block_position_data["block_entity_data"])
						}
						// 拿一下这个方块的方块实体数据
						if foreground_blockName == "chest" || foreground_blockName == "trapped_chest" {
							var useOfChest string = "chest"
							if foreground_blockName == "chest" {
								useOfChest = "trapped_chest"
							}
							// 如果这是个箱子，那么先放个陷阱箱
							// 反过来，如果这是个陷阱箱，那么先放个箱子
							ans = append(ans, &types.Module{
								Block: &types.Block{
									Name: &useOfChest,
									Data: 0,
								},
								Point: types.Position{
									X: int(i[key].BeginX - currentExport.BeginX),
									Y: int(i[key].BeginY + j - currentExport.BeginY),
									Z: int(i[key].BeginZ - currentExport.BeginZ),
								},
							})
						}
						// 对于箱子和陷阱箱的附加处理是为了解决箱子间的连接问题，让所有的箱子都不再连接
						// 不知道有没有人愿意解决这个问题呢？
						hasNBT = true
						blockNBT, err = nbt.MarshalEncoding(block_entity_data, nbt.LittleEndian)
						if err != nil {
							return []*types.Module{}, fmt.Errorf("DumpBlocks: %v", err)
						}
						// 取得 nbt
					}
				}
				// 取得方块实体数据
				if foreground_blockName != "undefined" && (background_blockName == "water" || background_blockName == "flowing_water") {
					ans = append(ans, &types.Module{
						Block: &types.Block{
							Name:        &background_blockName,
							BlockStates: background_blockStates,
						},
						Point: types.Position{
							X: int(i[key].BeginX - currentExport.BeginX),
							Y: int(i[key].BeginY + j - currentExport.BeginY),
							Z: int(i[key].BeginZ - currentExport.BeginZ),
						},
					})
				}
				// 含水类方块的处理。
				// 我们将含水类方块处理为 setblock water + targetBlock 的形式
				if foreground_blockName != "air" && foreground_blockName != "undefined" {
					single := &types.Module{
						Block: &types.Block{
							Name: &foreground_blockName,
						},
						Point: types.Position{
							X: int(i[key].BeginX - currentExport.BeginX),
							Y: int(i[key].BeginY + j - currentExport.BeginY),
							Z: int(i[key].BeginZ - currentExport.BeginZ),
						},
					}
					// 初始化单个元素
					if hasNBT {
						single.NBTData = blockNBT
					}
					single.Block.BlockStates = foreground_blockStates
					// 放入数据
					ans = append(ans, single)
					// 提交
				}
				// 放置前景层的方块
			}
		}
	}
	return ans, nil
}
