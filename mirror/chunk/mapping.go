package chunk

import (
	_ "embed"
	"phoenixbuilder/mirror/blocks"

	"github.com/lucasb-eyer/go-colorful"
)

var AirRID uint32 = blocks.AIR_RUNTIMEID

var (
	// SubChunkVersion is the current version of the written sub chunks, specifying the format they are
	// written on disk and over network.
	SubChunkVersion = 9
	// CurrentBlockVersion is the current version of blocks (states) of the game. This version is composed
	// of 4 bytes indicating a version, interpreted as a big endian int. The current version represents
	// 1.16.0.14 {1, 16, 0, 14}.
	// 1.19.10.22 !!
	CurrentBlockVersion int32 = int32(blocks.NEMC_BLOCK_VERSION)
)

// blockEntry represents a block as found in a disk save of a world.
type blockEntry struct {
	Name    string                 `nbt:"name"`
	State   map[string]interface{} `nbt:"states"`
	Version int32                  `nbt:"version"`
	ID      int32                  `nbt:"oldid,omitempty"` // PM writes this field, so we allow it anyway to avoid issues loading PM worlds.
	Meta    int16                  `nbt:"val,omitempty"`
}

type LegacyBlock struct {
	Name string
	Val  uint16
}

type ColorBlock struct {
	Color colorful.Color
	Block *LegacyBlock
}

var ColorTable = []ColorBlock{
	{Block: &LegacyBlock{Name: "stone", Val: 0}, Color: colorful.Color{89, 89, 89}},
	{Block: &LegacyBlock{Name: "stone", Val: 1}, Color: colorful.Color{135, 102, 76}},
	{Block: &LegacyBlock{Name: "stone", Val: 3}, Color: colorful.Color{237, 235, 229}},
	{Block: &LegacyBlock{Name: "stone", Val: 5}, Color: colorful.Color{104, 104, 104}},
	{Block: &LegacyBlock{Name: "grass", Val: 0}, Color: colorful.Color{144, 174, 94}},
	{Block: &LegacyBlock{Name: "planks", Val: 0}, Color: colorful.Color{129, 112, 73}},
	{Block: &LegacyBlock{Name: "planks", Val: 1}, Color: colorful.Color{114, 81, 51}},
	{Block: &LegacyBlock{Name: "planks", Val: 2}, Color: colorful.Color{228, 217, 159}},
	{Block: &LegacyBlock{Name: "planks", Val: 4}, Color: colorful.Color{71, 71, 71}},
	{Block: &LegacyBlock{Name: "planks", Val: 5}, Color: colorful.Color{91, 72, 50}},
	{Block: &LegacyBlock{Name: "leaves", Val: 0}, Color: colorful.Color{64, 85, 32}},
	{Block: &LegacyBlock{Name: "leaves", Val: 1}, Color: colorful.Color{54, 75, 50}},
	{Block: &LegacyBlock{Name: "leaves", Val: 2}, Color: colorful.Color{68, 83, 47}},
	{Block: &LegacyBlock{Name: "leaves", Val: 14}, Color: colorful.Color{58, 71, 40}},
	{Block: &LegacyBlock{Name: "leaves", Val: 15}, Color: colorful.Color{55, 73, 28}},
	{Block: &LegacyBlock{Name: "sponge", Val: 0}, Color: colorful.Color{183, 183, 70}},
	{Block: &LegacyBlock{Name: "lapis_block", Val: 0}, Color: colorful.Color{69, 101, 198}},
	{Block: &LegacyBlock{Name: "noteblock", Val: 0}, Color: colorful.Color{111, 95, 63}},
	{Block: &LegacyBlock{Name: "web", Val: 0}, Color: colorful.Color{159, 159, 159}},
	{Block: &LegacyBlock{Name: "wool", Val: 0}, Color: colorful.Color{205, 205, 205}},
	{Block: &LegacyBlock{Name: "wool", Val: 1}, Color: colorful.Color{163, 104, 54}},
	{Block: &LegacyBlock{Name: "wool", Val: 2}, Color: colorful.Color{132, 65, 167}},
	{Block: &LegacyBlock{Name: "wool", Val: 3}, Color: colorful.Color{91, 122, 169}},
	{Block: &LegacyBlock{Name: "wool", Val: 5}, Color: colorful.Color{115, 162, 53}},
	{Block: &LegacyBlock{Name: "wool", Val: 6}, Color: colorful.Color{182, 106, 131}},
	{Block: &LegacyBlock{Name: "wool", Val: 7}, Color: colorful.Color{60, 60, 60}},
	{Block: &LegacyBlock{Name: "wool", Val: 8}, Color: colorful.Color{123, 123, 123}},
	{Block: &LegacyBlock{Name: "wool", Val: 9}, Color: colorful.Color{69, 100, 121}},
	{Block: &LegacyBlock{Name: "wool", Val: 10}, Color: colorful.Color{94, 52, 137}},
	{Block: &LegacyBlock{Name: "wool", Val: 11}, Color: colorful.Color{45, 59, 137}},
	{Block: &LegacyBlock{Name: "wool", Val: 12}, Color: colorful.Color{78, 61, 43}},
	{Block: &LegacyBlock{Name: "wool", Val: 13}, Color: colorful.Color{85, 100, 49}},
	{Block: &LegacyBlock{Name: "wool", Val: 14}, Color: colorful.Color{113, 46, 44}},
	{Block: &LegacyBlock{Name: "wool", Val: 15}, Color: colorful.Color{20, 20, 20}},
	{Block: &LegacyBlock{Name: "gold_block", Val: 0}, Color: colorful.Color{198, 191, 84}},
	{Block: &LegacyBlock{Name: "iron_block", Val: 0}, Color: colorful.Color{134, 134, 134}},
	{Block: &LegacyBlock{Name: "double_stone_slab", Val: 1}, Color: colorful.Color{196, 187, 136}},
	{Block: &LegacyBlock{Name: "double_stone_slab", Val: 6}, Color: colorful.Color{204, 202, 196}},
	{Block: &LegacyBlock{Name: "double_stone_slab", Val: 7}, Color: colorful.Color{81, 11, 5}},
	{Block: &LegacyBlock{Name: "redstone_block", Val: 0}, Color: colorful.Color{188, 39, 26}},
	{Block: &LegacyBlock{Name: "mossy_cobblestone", Val: 0}, Color: colorful.Color{131, 134, 146}},
	{Block: &LegacyBlock{Name: "diamond_block", Val: 0}, Color: colorful.Color{102, 173, 169}},
	{Block: &LegacyBlock{Name: "farmland", Val: 0}, Color: colorful.Color{116, 88, 65}},
	{Block: &LegacyBlock{Name: "ice", Val: 0}, Color: colorful.Color{149, 149, 231}},
	{Block: &LegacyBlock{Name: "pumpkin", Val: 0}, Color: colorful.Color{189, 122, 62}},
	{Block: &LegacyBlock{Name: "monster_egg", Val: 1}, Color: colorful.Color{153, 156, 169}},
	{Block: &LegacyBlock{Name: "red_mushroom_block", Val: 0}, Color: colorful.Color{131, 53, 50}},
	// {Block: &LegacyBlock{Name: "vine", Val: 1}, Color: colorful.Color{68, 89, 34}},
	{Block: &LegacyBlock{Name: "brewing_stand", Val: 6}, Color: colorful.Color{155, 155, 155}},
	{Block: &LegacyBlock{Name: "double_wooden_slab", Val: 1}, Color: colorful.Color{98, 70, 44}},
	{Block: &LegacyBlock{Name: "emerald_block", Val: 0}, Color: colorful.Color{77, 171, 67}},
	{Block: &LegacyBlock{Name: "raw_gold_block", Val: 0}, Color: colorful.Color{231, 221, 99}},
	{Block: &LegacyBlock{Name: "stained_hardened_clay", Val: 0}, Color: colorful.Color{237, 237, 237}},
	{Block: &LegacyBlock{Name: "stained_hardened_clay", Val: 2}, Color: colorful.Color{154, 76, 194}},
	{Block: &LegacyBlock{Name: "stained_hardened_clay", Val: 4}, Color: colorful.Color{213, 213, 82}},
	{Block: &LegacyBlock{Name: "stained_hardened_clay", Val: 6}, Color: colorful.Color{211, 123, 153}},
	{Block: &LegacyBlock{Name: "stained_hardened_clay", Val: 8}, Color: colorful.Color{142, 142, 142}},
	{Block: &LegacyBlock{Name: "stained_hardened_clay", Val: 10}, Color: colorful.Color{110, 62, 160}},
	{Block: &LegacyBlock{Name: "slime", Val: 0}, Color: colorful.Color{109, 141, 60}},
	{Block: &LegacyBlock{Name: "packed_ice", Val: 0}, Color: colorful.Color{128, 128, 199}},
	{Block: &LegacyBlock{Name: "repeating_command_block", Val: 1}, Color: colorful.Color{77, 43, 112}},
	{Block: &LegacyBlock{Name: "chain_command_block", Val: 1}, Color: colorful.Color{70, 82, 40}},
	{Block: &LegacyBlock{Name: "nether_wart_block", Val: 0}, Color: colorful.Color{93, 38, 36}},
	{Block: &LegacyBlock{Name: "bone_block", Val: 0}, Color: colorful.Color{160, 153, 112}},
}
