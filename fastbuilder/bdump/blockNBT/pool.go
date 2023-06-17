package blockNBT

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
	"cauldron":           "Container",
	"lava_cauldron":      "Container",
	"jukebox":            "Container",
	"brewing_stand":      "Container",
	"undyed_shulker_box": "Container",
	"shulker_box":        "Container",
	"lectern":            "Container",
	// 容器
	"standing_sign":          "Sign",
	"spruce_standing_sign":   "Sign",
	"birch_standing_sign":    "Sign",
	"jungle_standing_sign":   "Sign",
	"acacia_standing_sign":   "Sign",
	"darkoak_standing_sign":  "Sign",
	"mangrove_standing_sign": "Sign",
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
	"bamboo_wall_sign":       "Sign",
	"crimson_wall_sign":      "Sign",
	"warped_wall_sign":       "Sign",
	"sign":                   "Sign",
	"spruce_sign":            "Sign",
	"birch_sign":             "Sign",
	"jungle_sign":            "Sign",
	"acacia_sign":            "Sign",
	"darkoak_sign":           "Sign",
	"mangrove_sign":          "Sign",
	"bamboo_sign":            "Sign",
	"crimson_sign":           "Sign",
	"warped_sign":            "Sign",
	"oak_hanging_sign":       "Sign",
	"spruce_hanging_sign":    "Sign",
	"birch_hanging_sign":     "Sign",
	"jungle_hanging_sign":    "Sign",
	"acacia_hanging_sign":    "Sign",
	"dark_oak_hanging_sign":  "Sign",
	"mangrove_hanging_sign":  "Sign",
	"bamboo_hanging_sign":    "Sign",
	"crimson_hanging_sign":   "Sign",
	"warped_hanging_sign":    "Sign",
	// 告示牌
}

// 用于放置方块实体的通用接口
type GeneralBlockNBT interface {
	Decode() error     // 解析 NBT(map[string]interface{}) 数据为 GoLang Struct
	WriteDatas() error // 放置方块并以正常玩家的行为写入方块实体数据
}

// 检查这个方块实体是否已被支持。
// 如果尚未被支持，则返回空字符串，否则返回这种方块的类型。
// 以告示牌为例，所有的告示牌都可以写作为 Sign
func CheckIfIsEffectiveNBTBlock(blockName string) string {
	value, ok := SupportBlocksPool[blockName]
	if ok {
		return value
	}
	return ""
}

// 取得用于放置目标方块实体的 接口/方法
func GetMethod(pack Package) GeneralBlockNBT {
	switch pack.Datas.Type {
	case "CommandBlock":
		return &CommandBlock{Package: &pack, NeedToPlaceBlock: true}
		// 命令方块
	case "Container":
		return &Container{Package: &pack}
		// 容器
	case "Sign":
		return &Sign{Package: &pack}
		// 告示牌
	default:
		return &Default{Package: &pack}
		// 其他尚且未被支持的方块实体
	}
	// 返回值
}
