package NBTAssigner

import GameInterface "phoenixbuilder/game_control/game_interface"

// 此表描述了现阶段已经支持了的方块实体。
// 键代表方块名，而值代表这种方块应该归属的类型
var SupportBlocksPool map[string]string = map[string]string{
	"command_block":           "CommandBlock",
	"chain_command_block":     "CommandBlock",
	"repeating_command_block": "CommandBlock",
	// 命令方块
	"blast_furnace":          "Container",
	"lit_blast_furnace":      "Container",
	"smoker":                 "Container",
	"lit_smoker":             "Container",
	"furnace":                "Container",
	"lit_furnace":            "Container",
	"chest":                  "Container",
	"barrel":                 "Container",
	"trapped_chest":          "Container",
	"hopper":                 "Container",
	"dispenser":              "Container",
	"dropper":                "Container",
	"jukebox":                "Container",
	"brewing_stand":          "Container",
	"undyed_shulker_box":     "Container",
	"shulker_box":            "Container", // 已在当前版本弃用
	"white_shulker_box":      "Container",
	"orange_shulker_box":     "Container",
	"magenta_shulker_box":    "Container",
	"light_blue_shulker_box": "Container",
	"yellow_shulker_box":     "Container",
	"lime_shulker_box":       "Container",
	"pink_shulker_box":       "Container",
	"gray_shulker_box":       "Container",
	"light_gray_shulker_box": "Container",
	"cyan_shulker_box":       "Container",
	"purple_shulker_box":     "Container",
	"blue_shulker_box":       "Container",
	"brown_shulker_box":      "Container",
	"green_shulker_box":      "Container",
	"red_shulker_box":        "Container",
	"black_shulker_box":      "Container",
	"lectern":                "Container",
	// 容器
	"standing_sign":          "Sign",
	"spruce_standing_sign":   "Sign",
	"birch_standing_sign":    "Sign",
	"jungle_standing_sign":   "Sign",
	"acacia_standing_sign":   "Sign",
	"darkoak_standing_sign":  "Sign",
	"mangrove_standing_sign": "Sign",
	"cherry_standing_sign":   "Sign",
	"bamboo_standing_sign":   "Sign",
	"crimson_standing_sign":  "Sign",
	"warped_standing_sign":   "Sign",
	"wall_sign":              "Sign",
	"spruce_wall_sign":       "Sign",
	"birch_wall_sign":        "Sign",
	"jungle_wall_sign":       "Sign",
	"acacia_wall_sign":       "Sign",
	"darkoak_wall_sign":      "Sign",
	"mangrove_wall_sign":     "Sign",
	"cherry_wall_sign":       "Sign",
	"bamboo_wall_sign":       "Sign",
	"crimson_wall_sign":      "Sign",
	"warped_wall_sign":       "Sign",
	"oak_hanging_sign":       "Sign",
	"spruce_hanging_sign":    "Sign",
	"birch_hanging_sign":     "Sign",
	"jungle_hanging_sign":    "Sign",
	"acacia_hanging_sign":    "Sign",
	"dark_oak_hanging_sign":  "Sign",
	"mangrove_hanging_sign":  "Sign",
	"cherry_hanging_sign":    "Sign",
	"bamboo_hanging_sign":    "Sign",
	"crimson_hanging_sign":   "Sign",
	"warped_hanging_sign":    "Sign",
	// 告示牌
}

// 此表描述了现阶段已经支持了的特殊物品，如烟花等物品。
// 键代表物品名，而值代表这种物品应该归属的类型
var SupportItemsPool map[string]string = map[string]string{
	"writable_book": "Book",
	"written_book":  "Book",
	// 成书
}

// 此表描述了现阶段已支持的方块实体中，
// 其物品名称到方块名的映射，
// 这将仅被用于子方块功能。
// 键代表物品名，而值代表此物品对应的方块名
var ItemNameToBlockNamePool map[string]string = map[string]string{
	"command_block":           "command_block",
	"chain_command_block":     "chain_command_block",
	"repeating_command_block": "repeating_command_block",
	// 命令方块
	"blast_furnace":          "blast_furnace",
	"lit_blast_furnace":      "lit_blast_furnace",
	"smoker":                 "smoker",
	"lit_smoker":             "lit_smoker",
	"furnace":                "furnace",
	"lit_furnace":            "lit_furnace",
	"chest":                  "chest",
	"barrel":                 "barrel",
	"trapped_chest":          "trapped_chest",
	"hopper":                 "hopper",
	"dispenser":              "dispenser",
	"dropper":                "dropper",
	"jukebox":                "jukebox",
	"brewing_stand":          "brewing_stand",
	"undyed_shulker_box":     "undyed_shulker_box",
	"shulker_box":            "shulker_box", // 已在当前版本弃用
	"white_shulker_box":      "white_shulker_box",
	"orange_shulker_box":     "orange_shulker_box",
	"magenta_shulker_box":    "magenta_shulker_box",
	"light_blue_shulker_box": "light_blue_shulker_box",
	"yellow_shulker_box":     "yellow_shulker_box",
	"lime_shulker_box":       "lime_shulker_box",
	"pink_shulker_box":       "pink_shulker_box",
	"gray_shulker_box":       "gray_shulker_box",
	"light_gray_shulker_box": "light_gray_shulker_box",
	"cyan_shulker_box":       "cyan_shulker_box",
	"purple_shulker_box":     "purple_shulker_box",
	"blue_shulker_box":       "blue_shulker_box",
	"brown_shulker_box":      "brown_shulker_box",
	"green_shulker_box":      "green_shulker_box",
	"red_shulker_box":        "red_shulker_box",
	"black_shulker_box":      "black_shulker_box",
	"lectern":                "lectern",
	// 容器
	"oak_sign":      "wall_sign",
	"spruce_sign":   "spruce_wall_sign",
	"birch_sign":    "birch_wall_sign",
	"jungle_sign":   "jungle_wall_sign",
	"acacia_sign":   "acacia_wall_sign",
	"darkoak_sign":  "darkoak_wall_sign",
	"mangrove_sign": "mangrove_wall_sign",
	"cherry_sign":   "cherry_wall_sign",
	"bamboo_sign":   "bamboo_wall_sign",
	"crimson_sign":  "crimson_wall_sign",
	"warped_sign":   "warped_wall_sign",
	// 告示牌
}

// 此表描述了可被 replaceitem 生效的容器
var SupportContainerPool map[string]ContainerConstantData = map[string]ContainerConstantData{
	"blast_furnace": {
		StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDInvalid,
		ContainerIDMapping: map[uint8]uint8{
			0: GameInterface.ContainerIDBlastFurnaceIngredient,
			1: GameInterface.ContainerIDBlastFurnaceFuel,
			2: GameInterface.ContainerIDBlastFurnaceResult,
		},
	},
	"lit_blast_furnace": {
		StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDInvalid,
		ContainerIDMapping: map[uint8]uint8{
			0: GameInterface.ContainerIDBlastFurnaceIngredient,
			1: GameInterface.ContainerIDBlastFurnaceFuel,
			2: GameInterface.ContainerIDBlastFurnaceResult,
		},
	},
	// 高炉(发光的高炉)
	"smoker": {
		StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDInvalid,
		ContainerIDMapping: map[uint8]uint8{
			0: GameInterface.ContainerIDSmokerIngredient,
			1: GameInterface.ContainerIDSmokerFuel,
			2: GameInterface.ContainerIDSmokerResult,
		},
	},
	"lit_smoker": {
		StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDInvalid,
		ContainerIDMapping: map[uint8]uint8{
			0: GameInterface.ContainerIDSmokerIngredient,
			1: GameInterface.ContainerIDSmokerFuel,
			2: GameInterface.ContainerIDSmokerResult,
		},
	},
	// 烟熏炉(发光的烟熏炉)
	"furnace": {
		StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDInvalid,
		ContainerIDMapping: map[uint8]uint8{
			0: GameInterface.ContainerIDFurnaceIngredient,
			1: GameInterface.ContainerIDFurnaceFuel,
			2: GameInterface.ContainerIDFurnaceResult,
		},
	},
	"lit_furnace": {
		StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDInvalid,
		ContainerIDMapping: map[uint8]uint8{
			0: GameInterface.ContainerIDFurnaceIngredient,
			1: GameInterface.ContainerIDFurnaceFuel,
			2: GameInterface.ContainerIDFurnaceResult,
		},
	},
	// 熔炉(发光的熔炉)
	"brewing_stand": {
		StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDInvalid,
		ContainerIDMapping: map[uint8]uint8{
			0: GameInterface.ContainerIDBrewingStandInput,
			1: GameInterface.ContainerIDBrewingStandResult,
			2: GameInterface.ContainerIDBrewingStandResult,
			3: GameInterface.ContainerIDBrewingStandResult,
			4: GameInterface.ContainerIDBrewingStandFuel,
		},
	},
	// 酿造台
	"chest":         {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDDefault},
	"trapped_chest": {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDDefault},
	"lectern":       {StorageItemValue: "book", CouldOpen: false, UniversalContainerID: GameInterface.ContainerIDUnknown},
	"hopper":        {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDDefault},
	"dispenser":     {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDDefault},
	"dropper":       {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDDefault},
	"jukebox":       {StorageItemValue: "RecordItem", CouldOpen: false, UniversalContainerID: GameInterface.ContainerIDUnknown},
	"barrel":        {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDBarrel},
	// 箱子、陷阱箱、讲台、漏洞、发射器、投掷器、唱片机、木桶
	"undyed_shulker_box":     {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"shulker_box":            {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox}, // 已在当前版本弃用
	"white_shulker_box":      {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"orange_shulker_box":     {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"magenta_shulker_box":    {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"light_blue_shulker_box": {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"yellow_shulker_box":     {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"lime_shulker_box":       {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"pink_shulker_box":       {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"gray_shulker_box":       {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"light_gray_shulker_box": {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"cyan_shulker_box":       {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"purple_shulker_box":     {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"blue_shulker_box":       {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"brown_shulker_box":      {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"green_shulker_box":      {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"red_shulker_box":        {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	"black_shulker_box":      {StorageItemValue: "Items", CouldOpen: true, UniversalContainerID: GameInterface.ContainerIDShulkerBox},
	// 各种颜色或未被染色的潜影盒
}

// 此表描述了染料 RGB 颜色值到 染料物品名 的映射
var RGBToDyeItemName map[[3]uint8]string = map[[3]uint8]string{
	{240, 240, 240}: "white_dye",      // 白色染料
	{157, 151, 151}: "light_gray_dye", // 淡灰色染料
	{71, 79, 82}:    "gray_dye",       // 灰色染料
	{0, 0, 0}:       "black_dye",      // 黑色染料
	{131, 84, 50}:   "brown_dye",      // 棕色染料
	{176, 46, 38}:   "red_dye",        // 红色染料
	{249, 128, 29}:  "orange_dye",     // 橙色染料
	{254, 216, 61}:  "yellow_dye",     // 黄色染料
	{128, 199, 31}:  "lime_dye",       // 黄绿色染料
	{94, 124, 22}:   "green_dye",      // 绿色染料
	{22, 156, 156}:  "cyan_dye",       // 青色染料
	{58, 179, 218}:  "light_blue_dye", // 淡蓝色染料
	{60, 68, 170}:   "blue_dye",       // 蓝色染料
	{137, 50, 184}:  "purple_dye",     // 紫色染料
	{199, 78, 189}:  "magenta_dye",    // 品红色染料
	{243, 139, 170}: "pink_dye",       // 粉红色染料
}

// 此表描述了 MCBE 所有原本染料的 RGB 颜色
var DefaultDyeColor [][3]uint8 = [][3]uint8{
	{240, 240, 240}, // 白色
	{157, 151, 151}, // 淡灰色
	{71, 79, 82},    // 灰色
	{0, 0, 0},       // 黑色(告示牌默认颜色)
	{131, 84, 50},   // 棕色
	{176, 46, 38},   // 红色
	{249, 128, 29},  // 橙色
	{254, 216, 61},  // 黄色
	{128, 199, 31},  // 黄绿色
	{94, 124, 22},   // 绿色
	{22, 156, 156},  // 青色
	{58, 179, 218},  // 淡蓝色
	{60, 68, 170},   // 蓝色
	{137, 50, 184},  // 紫色
	{199, 78, 189},  // 品红色
	{243, 139, 170}, // 粉红色
}
