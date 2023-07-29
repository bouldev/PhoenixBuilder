package NBTAssigner

// 此表描述了现阶段已经支持了的方块实体。
// 键代表方块名，而值代表这种方块应该归属的类型
var SupportBlocksPool map[string]string = map[string]string{
	"command_block":           "CommandBlock",
	"chain_command_block":     "CommandBlock",
	"repeating_command_block": "CommandBlock",
	// 命令方块
	"blast_furnace":      "Container",
	"lit_blast_furnace":  "Container",
	"smoker":             "Container",
	"lit_smoker":         "Container",
	"furnace":            "Container",
	"lit_furnace":        "Container",
	"chest":              "Container",
	"barrel":             "Container",
	"trapped_chest":      "Container",
	"hopper":             "Container",
	"dispenser":          "Container",
	"dropper":            "Container",
	"jukebox":            "Container",
	"brewing_stand":      "Container",
	"undyed_shulker_box": "Container",
	"shulker_box":        "Container",
	"lectern":            "Container",
	// 容器
	"standing_sign":         "Sign",
	"spruce_standing_sign":  "Sign",
	"birch_standing_sign":   "Sign",
	"jungle_standing_sign":  "Sign",
	"acacia_standing_sign":  "Sign",
	"darkoak_standing_sign": "Sign",
	"crimson_standing_sign": "Sign",
	"warped_standing_sign":  "Sign",
	"wall_sign":             "Sign",
	"spruce_wall_sign":      "Sign",
	"birch_wall_sign":       "Sign",
	"jungle_wall_sign":      "Sign",
	"acacia_wall_sign":      "Sign",
	"darkoak_wall_sign":     "Sign",
	"crimson_wall_sign":     "Sign",
	"warped_wall_sign":      "Sign",
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
	"blast_furnace":      "blast_furnace",
	"lit_blast_furnace":  "lit_blast_furnace",
	"smoker":             "smoker",
	"lit_smoker":         "lit_smoker",
	"furnace":            "furnace",
	"lit_furnace":        "lit_furnace",
	"chest":              "chest",
	"barrel":             "barrel",
	"trapped_chest":      "trapped_chest",
	"hopper":             "hopper",
	"dispenser":          "dispenser",
	"dropper":            "dropper",
	"jukebox":            "jukebox",
	"brewing_stand":      "brewing_stand",
	"undyed_shulker_box": "undyed_shulker_box",
	"shulker_box":        "shulker_box",
	"lectern":            "lectern",
	// 容器
	"oak_sign":     "wall_sign",
	"spruce_sign":  "spruce_wall_sign",
	"birch_sign":   "birch_wall_sign",
	"jungle_sign":  "jungle_wall_sign",
	"acacia_sign":  "acacia_wall_sign",
	"darkoak_sign": "darkoak_wall_sign",
	"crimson_sign": "crimson_wall_sign",
	"warped_sign":  "warped_wall_sign",
	// 告示牌
}

// 此表描述了可被 replaceitem 生效的容器
var SupportContainerPool map[string]ContainerConstantData = map[string]ContainerConstantData{
	"blast_furnace":      {StorageItemValue: "Items", CouldOpen: true, ContainerID: 45},
	"lit_blast_furnace":  {StorageItemValue: "Items", CouldOpen: true, ContainerID: 45},
	"smoker":             {StorageItemValue: "Items", CouldOpen: true, ContainerID: 28},
	"lit_smoker":         {StorageItemValue: "Items", CouldOpen: true, ContainerID: 28},
	"furnace":            {StorageItemValue: "Items", CouldOpen: true, ContainerID: 25},
	"lit_furnace":        {StorageItemValue: "Items", CouldOpen: true, ContainerID: 25},
	"chest":              {StorageItemValue: "Items", CouldOpen: true, ContainerID: 7},
	"trapped_chest":      {StorageItemValue: "Items", CouldOpen: true, ContainerID: 7},
	"lectern":            {StorageItemValue: "book", CouldOpen: false, ContainerID: 255},
	"hopper":             {StorageItemValue: "Items", CouldOpen: true, ContainerID: 7},
	"dispenser":          {StorageItemValue: "Items", CouldOpen: true, ContainerID: 7},
	"dropper":            {StorageItemValue: "Items", CouldOpen: true, ContainerID: 7},
	"jukebox":            {StorageItemValue: "RecordItem", CouldOpen: false, ContainerID: 255},
	"brewing_stand":      {StorageItemValue: "Items", CouldOpen: true, ContainerID: 59},
	"undyed_shulker_box": {StorageItemValue: "Items", CouldOpen: true, ContainerID: 30},
	"shulker_box":        {StorageItemValue: "Items", CouldOpen: true, ContainerID: 30},
	// phoenixbuilder/game_control/game_interface/item_copier.go uses this field directly
	"barrel": {StorageItemValue: "Items", CouldOpen: true, ContainerID: 58},
}
