package lexport_depends

import (
	"fmt"
	"math"
	"phoenixbuilder/fastbuilder/types"
	TranslateNBTInerface "phoenixbuilder/io/special_tasks/lexport_depends/TranslateNBTInterface"
	"strconv"
	"strings"
)

// 用于描述一个区域的基本信息，也就是区域的起点位置及区域的尺寸
type Area struct {
	BeginX int
	BeginY int
	BeginZ int
	SizeX  int
	SizeY  int
	SizeZ  int
}

// 用于描述一个区域的坐标
type AreaLocation struct {
	Posx int
	Posz int
}

// 用于描述一个方块的坐标
type BlockPos struct {
	Posx int
	Posy int
	Posz int
}

/*
用于存放一个 MCBE 的结构；这里面的数据稍微作了一些处理，只保留了需要的部分

如果后期要给这个结构体添加别的东西，请参见本文件中的 GetMCStructureData 函数
*/
type Mcstructure struct {
	info                     Area                           // 用于描述这个结构的基本信息，也就是起点位置及尺寸
	blockPalette             []string                       // 用于存放调色板(方块池)中的方块名
	blockPalette_blockStates []string                       // 用于存放调色板(方块池)中的数据；这里的方块池稍作了处理，只保留了方块状态(string)，且这种方块状态正是 setblock 命令所需要的部分；需要特别说明的是，方块状态里面所有的 TAG_Byte 都被处理成了布尔值，如果有 BUG 记得提 Issue
	blockPalette_blockData   []int16                        // 用于存放调色板(方块池)中的数据；这里的方块池稍作了处理，只保留了方块数据值，也就是附加值(int)；这个东西只是为了支持容器而做的
	foreground               []int16                        // 用于描述一个方块的前景层；这里应该用 int32 的，不过 PhoenixBuilder 只能表示 int16 个方块，所以我这里就省一下内存
	background               []int16                        // 用于描述一个方块的背景层；这里应该用 int32 的，不过 PhoenixBuilder 只能表示 int16 个方块，所以我这里就省一下内存
	blockNBT                 map[int]map[string]interface{} // 用于存放方块实体数据
}

/*
用于拆分一个大区域为若干个小区域；当 useSpecialSplitWay 为真时，将蛇形拆分区域

返回值 []Area 代表一个已经排好顺序的若干个小区域

返回值 map[AreaLocation]int 代表可以通过 区域坐标(AreaLocation) 来访问 []Area 的对应项

因此，返回值 map[int]AreaLocation 是返回值 map[AreaLocation]int 的逆过程
*/
func SplitArea(
	startX int, startY int, startZ int,
	endX int, endY int, endZ int,
	splitSizeX int, splitSizeZ int,
	useSpecialSplitWay bool,
) ([]Area, map[AreaLocation]int, map[int]AreaLocation) {
	if splitSizeX < 0 {
		splitSizeX = splitSizeX * -1
	}
	if splitSizeZ < 0 {
		splitSizeZ = splitSizeZ * -1
	}
	// 考虑一些特殊的情况，此举是为了更高的兼容性
	var save int
	if endX < startX {
		save = startX
		startX = endX
		endX = save
	}
	if endY < startY {
		save = startY
		startY = endY
		endY = save
	}
	if endZ < startZ {
		save = startZ
		startZ = endZ
		endZ = save
	}
	// 考虑一些特殊的情况，此举是为了更高的兼容性
	sizeX := endX - startX + 1
	sizeY := endY - startY + 1
	sizeZ := endZ - startZ + 1
	// 取得 Area 的大小
	chunkX_length := int(math.Ceil(float64(sizeX) / float64(splitSizeX)))
	chunkZ_length := int(math.Ceil(float64(sizeZ) / float64(splitSizeZ)))
	// 取得各轴上需要拆分的区域数
	ans := make([]Area, chunkX_length*chunkZ_length) // 这个东西最终会 return 掉
	areaLoctionToInt := map[AreaLocation]int{}       // 知道了区域的坐标求区域在 []Area 的位置
	IntToareaLoction := map[int]AreaLocation{}       // 知道了区域在 []Area 的位置求区域坐标
	facing := -1                                     // 蛇形处理的时候需要用到这个
	key := -1                                        // 向 ans 插入数据的时候需要用到这个
	// 初始化
	for chunkX := 1; chunkX <= chunkX_length; chunkX++ {
		facing = facing * -1
		BeginX := splitSizeX*(chunkX-1) + startX
		xLength := splitSizeX
		if BeginX+xLength-1 > endX {
			xLength = endX - BeginX + 1
		}
		for chunkZ := 1; chunkZ <= chunkZ_length; chunkZ++ {
			key++ // p = p + 1
			currentChunkZ := chunkZ
			if useSpecialSplitWay && facing == -1 {
				currentChunkZ = chunkZ_length - currentChunkZ + 1
			}
			BeginZ := splitSizeZ*(currentChunkZ-1) + startZ
			zLength := splitSizeZ
			if BeginZ+zLength-1 > endZ {
				zLength = endZ - BeginZ + 1
			}
			ans[key] = Area{
				BeginX: BeginX,
				BeginY: startY,
				BeginZ: BeginZ,
				SizeX:  xLength,
				SizeY:  sizeY,
				SizeZ:  zLength,
			}
			areaLoctionToInt[AreaLocation{chunkX - 1, currentChunkZ - 1}] = key
			IntToareaLoction[key] = AreaLocation{chunkX - 1, currentChunkZ - 1}
		}
	}
	return ans, areaLoctionToInt, IntToareaLoction
}

// 用于提取得到的 MCBE 结构文件中的一些数据，具体拿了什么数据，你可以看返回值字段
func GetMCStructureData(area Area, structure map[string]interface{}) (Mcstructure, error) {
	var value_default map[string]interface{} = map[string]interface{}{}
	var ok bool = false
	var normal = false

	var value_structure map[string]interface{} = map[string]interface{}{}

	var blockPalette = []string{}
	var blockPalette_blockStates []string = []string{}
	var blockPalette_blockData []int16 = []int16{}
	var blockNBT map[int]map[string]interface{} = map[int]map[string]interface{}{}
	var foreground []int16 = []int16{}
	var background []int16 = []int16{}
	// 初始化
	_, ok = structure["structure"]
	if !ok {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"]")
	}
	value_structure, normal = structure["structure"].(map[string]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"]")
	}
	// input["structure"]
	_, ok = value_structure["palette"]
	if !ok {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"]")
	}
	value_palette, normal := value_structure["palette"].(map[string]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"]")
	}
	// input["structure"]["palette"]
	_, ok = value_palette["default"]
	if !ok {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"]")
	}
	value_default, normal = value_palette["default"].(map[string]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"]")
	}
	// input["structure"]["palette"]["default"]
	_, ok = value_default["block_palette"]
	if !ok {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"]")
	}
	value_block_palette, normal := value_default["block_palette"].([]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"]")
	}
	// input["structure"]["palette"]["default"]["block_palette"]
	for key, value := range value_block_palette {
		got, normal := value.(map[string]interface{})
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"name\"]", key)
		}
		// 这里确认下数据类型，就是这个 got 必须得是个复合标签
		_, ok = got["name"]
		if !ok {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"name\"]", key)
		}
		value_name, normal := got["name"].(string)
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"name\"]", key)
		}
		blockPalette = append(blockPalette, value_name)
		// 得到方块的名称
		// 这里的名称是携带了命名空间 minecraft 的
		// 命名空间会在后边删掉
		_, ok = got["states"]
		if !ok {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"states\"]", key)
		}
		value_states, normal := got["states"].(map[string]interface{})
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"states\"]", key)
		}
		blockStates, err := TranslateNBTInerface.Compound(value_states, true)
		if err != nil {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"states\"]", key)
		}
		blockPalette_blockStates = append(blockPalette_blockStates, blockStates)
		// 得到方块的方块状态
		_, ok = got["val"]
		if !ok {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"val\"]", key)
		}
		val, normal := got["val"].(int16)
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"palette\"][\"default\"][\"block_palette\"][%v][\"val\"]", key)
		}
		blockPalette_blockData = append(blockPalette_blockData, val)
		// 得到方块的方块数据值(附加值)
	}
	// 这个 for 用于获取调色板的信息的
	_, ok = value_default["block_position_data"]
	if ok {
		value_block_position_data, normal := value_default["block_position_data"].(map[string]interface{})
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"default\"][\"block_position_data\"]")
		}
		for key, value := range value_block_position_data {
			block_position_data, ok := value.(map[string]interface{})
			if !ok {
				return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"default\"][\"block_position_data\"][%v]", key)
			}
			location_of_block_position_data, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"default\"][\"block_position_data\"][%v]", key)
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
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"block_indices\"]")
	}
	value_block_indices, normal := value_structure["block_indices"].([]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"block_indices\"]")
	}
	// input["structure"]["block_indices"]
	if len(value_block_indices) != 2 {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"block_indices\"]")
	}
	// 这里要求 input["structure"]["block_indices"] 的长度必须为 2
	// 毕竟是由 前景层 和 背景层 的索引所制成的两张表
	value_block_indices_0, normal := value_block_indices[0].([]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"block_indices\"][0]")
	}
	for blockLocation_key, blockLocation := range value_block_indices_0 {
		got, normal := blockLocation.(int32)
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"block_indices\"][0][%v]", blockLocation_key)
		}
		foreground = append(foreground, int16(got))
	}
	// 这里先拿前景层方块的索引表
	value_block_indices_1, normal := value_block_indices[1].([]interface{})
	if !normal {
		return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"block_indices\"][1]")
	}
	for blockLocation_key, blockLocation := range value_block_indices_1 {
		got, normal := blockLocation.(int32)
		if !normal {
			return Mcstructure{}, fmt.Errorf("GetMCStructureData: Crashed in input[\"structure\"][\"block_indices\"][1][%v]", blockLocation_key)
		}
		background = append(background, int16(got))
	}
	// 然后再去拿背景层方块的索引表
	return Mcstructure{
		info:                     area,
		blockPalette:             blockPalette,
		blockPalette_blockStates: blockPalette_blockStates,
		blockPalette_blockData:   blockPalette_blockData,
		foreground:               foreground,
		background:               background,
		blockNBT:                 blockNBT,
	}, nil
	// 返回扒~
}

// 根据 mcstructure 的起点和尺寸，以及提供的方块坐标，寻找这个方块在 mcstructure 中的角标
func SearchForBlock(structureInfo Area, pos BlockPos) (int, error) {
	pos.Posx = pos.Posx - structureInfo.BeginX
	pos.Posy = pos.Posy - structureInfo.BeginY
	pos.Posz = pos.Posz - structureInfo.BeginZ
	// 将方块的绝对坐标转换为相对坐标(相对于 mcstructure)
	blockCount := structureInfo.SizeX * structureInfo.SizeY * structureInfo.SizeZ
	// 计算结构的尺寸
	angleMark := 0
	angleMark = angleMark + structureInfo.SizeY*structureInfo.SizeZ*pos.Posx
	angleMark = angleMark + structureInfo.SizeZ*pos.Posy
	angleMark = angleMark + pos.Posz
	// 计算方块相对于 mcstructure 的角标
	if angleMark > blockCount-1 {
		return -1, fmt.Errorf("Index out of the list, occured in input[%v]", angleMark)
	}
	return angleMark, nil
}

/*
基于区块的大小对整个待导出区域进行重排，并写入对应的方块、NBT数据

allAreas 对整个待导出区域按 64*64 大小拆分，且蛇形拆分(使用SplitArea拆分)，然后再获取拆分得到的各个小区域的 mcstructure 数据，然后处理后制成此 allAreas 表

allAreasFindUse 通过 区域坐标 来查这个区域在 allAreas 表的位置

currentExport 当前 Task 指定的导出区域，也就是根据 set(get) 和 setend(get end) 制成的 Area
*/
func ExportBaseOnChunkSize(
	allAreas []Mcstructure,
	allAreasFindUse map[AreaLocation]int,
	currentExport Area,
) ([]*types.Module, error) {
	ans := make([]*types.Module, 0)
	// 这个东西最后会 return 掉
	allChunks, _, allChunksFindUse := SplitArea(
		currentExport.BeginX, currentExport.BeginY, currentExport.BeginZ,
		currentExport.BeginX+currentExport.SizeX-1,
		currentExport.BeginY+currentExport.SizeY-1,
		currentExport.BeginZ+currentExport.SizeZ-1,
		16, 16, true,
	)
	// 将所有待导出区域按 16*16 的大小拆分为区块，且蛇形拆分
	// 然后按照得到的结果重排处理
	for key, value := range allChunks {
		chunkPos := allChunksFindUse[key]
		chunkPos.Posx = int(math.Floor(float64(chunkPos.Posx) / 4))
		chunkPos.Posz = int(math.Floor(float64(chunkPos.Posz) / 4))
		// 取得当前遍历的区块的坐标
		// 这里已经把坐标变换到 allAreas 下的坐标系中
		targetAreaPos := allAreasFindUse[chunkPos]
		targetArea := allAreas[targetAreaPos]
		// 取得被遍历区块对应的 mcstructure
		i, _, _ := SplitArea(
			value.BeginX, value.BeginY, value.BeginZ,
			value.BeginX+value.SizeX-1,
			value.BeginY+value.SizeY-1,
			value.BeginZ+value.SizeZ-1,
			1, 1, true,
		)
		allBlocksInCurrentChunk := make([]int32, 0)
		for _, VALUE := range i {
			got, err := SearchForBlock(targetArea.info, BlockPos{
				Posx: VALUE.BeginX,
				Posy: VALUE.BeginY,
				Posz: VALUE.BeginZ,
			})
			if err != nil {
				return []*types.Module{}, fmt.Errorf("SearchForBlock(Started by ExportBaseOnChunk): %v", err)
			}
			allBlocksInCurrentChunk = append(allBlocksInCurrentChunk, int32(got))
		}
		// 枚举出被遍历区块中所有方块的坐标(只枚举其中一层)
		for KEY, VALUE := range allBlocksInCurrentChunk {
			VALUE = VALUE - int32(targetArea.info.SizeZ)
			// 这个前置处理方法可能不太优雅
			// 凑合着用吧
			for j := 0; j < targetArea.info.SizeY; j++ {
				VALUE = VALUE + int32(targetArea.info.SizeZ)
				// 前往下一层
				foreground_blockName := "undefined"
				background_blockName := "undefined"
				foreground_blockStates := "undefined"
				background_blockStates := "undefined"
				foreground_blockData := int16(-1)
				// 初始化
				fgId := targetArea.foreground[VALUE] // 前景层方块在调色板中的id
				bgId := targetArea.background[VALUE] // 背景层方块在调色板中的id
				if fgId != -1 {
					foreground_blockName = strings.Replace(targetArea.blockPalette[fgId], "minecraft:", "", 1) // 前景层方块的名称
					foreground_blockStates = targetArea.blockPalette_blockStates[fgId]                         // 前景层方块的方块状态
					foreground_blockData = targetArea.blockPalette_blockData[fgId]                             // 前景层方块的方块数据值(附加值)
				}
				if bgId != -1 {
					background_blockName = strings.Replace(targetArea.blockPalette[bgId], "minecraft:", "", 1) // 背景层方块的名称
					background_blockStates = targetArea.blockPalette_blockStates[bgId]                         // 背景层方块的方块状态
				}
				// 获得基本信息
				var hasNBT bool = false
				var containerDataMark bool = false
				var containerData types.ChestData = types.ChestData{}
				var commandBlockDataMark bool = false
				var commandBlockData types.CommandBlockData = types.CommandBlockData{}
				var string_nbt string = ""
				var err error = fmt.Errorf("ExportBaseOnChunk: Initialization error")
				// 变量初始化
				// 危险！变量初始化这里不要动，不然可能会出现一些意想不到的 Bug
				got, ok := targetArea.blockNBT[int(VALUE)]
				if ok {
					_, ok := got["block_position_data"]
					if !ok {
						return []*types.Module{}, fmt.Errorf("ExportBaseOnChunk: Crashed by could not found \"block_position_data\"")
					}
					block_position_data, normal := got["block_position_data"].(map[string]interface{})
					if !normal {
						return []*types.Module{}, fmt.Errorf("ExportBaseOnChunk: Crashed by invalid \"block_position_data\"")
					}
					// 只要这个方块被记录了 NBT 数据，那么一定会有 block_position_data
					_, ok = block_position_data["block_entity_data"]
					// 虽然这个方块被记录了 NBT 数据，但不一定是一个方块实体
					// 所以即便这个被记录了 NBT 的方块没有 block_entity_data ，也不应该返回错误
					if ok {
						block_entity_data, normal := block_position_data["block_entity_data"].(map[string]interface{})
						if !normal {
							return []*types.Module{}, fmt.Errorf("ExportBaseOnChunk: Crashed by invalid \"block_entity_data\"")
						}
						// 拿一下这个方块的方块实体数据
						containerData, err = TranslateNBTInerface.GetContainerDataRun(block_entity_data, foreground_blockName)
						if fmt.Sprintf("%v", err) != "GetContainerDataRun: Not a container" && err != nil {
							return []*types.Module{}, fmt.Errorf("%v", err)
						}
						// 检查一下这个 NBT 方块是不是容器，如果不是会返回一个叫做 "GetContainerDataRun: Not a container" 的错误
						if err == nil {
							containerDataMark = true
							// 标记当前被处理的方块是一个容器
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
										X: i[KEY].BeginX - currentExport.BeginX,
										Y: i[KEY].BeginY + j - currentExport.BeginY,
										Z: i[KEY].BeginZ - currentExport.BeginZ,
									},
								})
							}
							// 对于箱子和陷阱箱的附加处理是为了解决箱子间的连接问题，让所有的箱子都不再连接；不知道有没有人愿意解决这个问题呢？
						}
						// 容器
						if foreground_blockName == "command_block" || foreground_blockName == "repeating_command_block" || foreground_blockName == "chain_command_block" {
							commandBlockData, err = TranslateNBTInerface.GetCommandBlockData(block_entity_data, foreground_blockName)
							if err != nil {
								return []*types.Module{}, fmt.Errorf("GetCommandBlockData(Started by ExportBaseOnChunk): %v", err)
							}
							commandBlockDataMark = true
						}
						// 命令方块
						hasNBT = true
						string_nbt, err = TranslateNBTInerface.Compound(block_entity_data, false)
						if err != nil {
							return []*types.Module{}, fmt.Errorf("%v", err)
						}
						// 取得 snbt
					}
				}
				// 取得方块实体数据
				if foreground_blockName != "" && foreground_blockName != "undefined" && (background_blockName == "water" || background_blockName == "flowing_water") {
					ans = append(ans, &types.Module{
						Block: &types.Block{
							Name:        &background_blockName,
							BlockStates: background_blockStates,
						},
						Point: types.Position{
							X: i[KEY].BeginX - currentExport.BeginX,
							Y: i[KEY].BeginY + j - currentExport.BeginY,
							Z: i[KEY].BeginZ - currentExport.BeginZ,
						},
					})
				}
				// 含水类方块
				// 我不清楚有没有其他“含”方块，有的话记得提醒我哦！
				// 这里处理的看似很拙劣，但实际上很有用！
				if foreground_blockName != "" && foreground_blockName != "air" && foreground_blockName != "undefined" {
					single := &types.Module{
						Block: &types.Block{
							Name: &foreground_blockName,
						},
						Point: types.Position{
							X: i[KEY].BeginX - currentExport.BeginX,
							Y: i[KEY].BeginY + j - currentExport.BeginY,
							Z: i[KEY].BeginZ - currentExport.BeginZ,
						},
					}
					// 简单地初始化一下一个单个的元素
					if commandBlockDataMark {
						single.Block.Data = uint16(foreground_blockData)
						single.CommandBlockData = &commandBlockData
					}
					// 命令方块
					if !commandBlockDataMark && containerDataMark {
						single.Block.Data = uint16(foreground_blockData)
						single.ChestData = &containerData
					}
					// 容器
					// 优先级比命令方块低一些
					if hasNBT {
						single.NBTData = []byte(string_nbt)
					}
					// operation 39 - RecordBlockEntityData
					// 更多信息请见
					// https://github.com/LNSSPsd/PhoenixBuilder/issues/83
					if !commandBlockDataMark && !containerDataMark {
						single.Block.BlockStates = foreground_blockStates
					}
					// 普通方块
					ans = append(ans, single)
					// 提交单个元素
				}
				// 放置前景层的方块
			}
		}
	}
	return ans, nil
}
