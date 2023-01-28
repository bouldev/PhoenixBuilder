package chunk

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"github.com/andybalholm/brotli"
	"github.com/lucasb-eyer/go-colorful"
)

// the input of this function is nemc runtime id, and should only from the network (chunk data packet), so it should never be "not found" if mapping and input is correct
var NEMCAirRID uint32
var AirRID uint32

// for query speed, these two functions do not have "found" return value
// for not found, it will return (mc/nemc) air rtid
// note: for an unknown (nemc/mc) rtid -> mapping -> (mc/nemc) air -> mapping -> (nemc/air) air, after twice conversion, it will be (nemc/mc) air, not the origin value
var NEMCRuntimeIDToStandardRuntimeID func(nemcRuntimeID uint32) (runtimeID uint32)
var StandardRuntimeIDToNEMCRuntimeID func(runtimeID uint32) (nemcRuntimeID uint32)

// the only place nemc runtime id is used is in decoding of chunk data packet, so in any other place, standard runtime id is used
// so in function below, the "runtimeID" is always standard runtime id, not nemc runtime id

var LegacyAirBlock *LegacyBlock
var JavaAirBlock = "minecraft:air"

// if not found, always return air block/rid/state, not nil/empty .etc
var RuntimeIDToLegacyBlock func(runtimeID uint32) (legacyBlock *LegacyBlock, found bool)
var LegacyBlockToRuntimeID func(name string, data uint16) (runtimeID uint32, found bool)
var StateToRuntimeID func(name string, properties map[string]any) (runtimeID uint32, found bool)
var RuntimeIDToState func(runtimeID uint32) (name string, properties map[string]any, found bool)
var RuntimeIDToBlock func(runtimeID uint32) (block *GeneralBlock, found bool)
var JavaToRuntimeID func(javaBlockStr string) (runtimeID uint32, found bool)
var RuntimeIDToJava func(runtimeID uint32) (javaBlockStr string, found bool)

var SchematicBlockToRuntimeID func(block, data byte) (runtimeID uint32, found bool)
var SchematicBlockToRuntimeIDStaticMapping []uint32
var stateRuntimeIDs = map[StateHash]uint32{}
var Blocks []*GeneralBlock
var LegacyBlocks []*LegacyBlock
var LegacyRuntimeIDs = map[LegacyBlockHash]uint32{}
var JavaStrToRuntimeIDMapping map[string]uint32
var JavaStrPropsToRuntimeIDMapping map[string]map[string]uint32
var RuntimeIDToJavaStrMapping map[uint32]string

const (
	// SubChunkVersion is the current version of the written sub chunks, specifying the format they are
	// written on disk and over network.
	SubChunkVersion = 9
	// CurrentBlockVersion is the current version of blocks (states) of the game. This version is composed
	// of 4 bytes indicating a version, interpreted as a big endian int. The current version represents
	// 1.16.0.14 {1, 16, 0, 14}.
	// 1.19.10.22 !!
	CurrentBlockVersion int32 = 18024982
)

// blockEntry represents a block as found in a disk save of a world.
type blockEntry struct {
	Name    string                 `nbt:"name"`
	State   map[string]interface{} `nbt:"states"`
	Version int32                  `nbt:"version"`
	ID      int32                  `nbt:"oldid,omitempty"` // PM writes this field, so we allow it anyway to avoid issues loading PM worlds.
	Meta    int16                  `nbt:"val,omitempty"`
}

type GeneralBlock struct {
	Name       string                 `nbt:"name"`
	Properties map[string]interface{} `nbt:"states"`
	Version    int32                  `nbt:"version"`
}

type LegacyBlock struct {
	Name string
	Val  uint16
}

func (b GeneralBlock) EncodeBlock() (string, map[string]any) {
	return b.Name, b.Properties
}

type StateHash struct {
	name, properties string
}

type LegacyBlockHash struct {
	name string
	data uint16
}

func hashProperties(properties map[string]interface{}) string {
	if properties == nil {
		return ""
	}
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	var b strings.Builder
	for _, k := range keys {
		switch v := properties[k].(type) {
		case bool:
			if v {
				b.WriteByte(1)
			} else {
				b.WriteByte(0)
			}
		case uint8:
			b.WriteByte(v)
		case int32:
			a := *(*[4]byte)(unsafe.Pointer(&v))
			b.Write(a[:])
		case string:
			b.WriteString(v)
		default:
			// If block encoding is broken, we want to find out as soon as possible. This saves a lot of time
			// debugging in-game.
			panic(fmt.Sprintf("invalid block property type %T for property %v", v, k))
		}
	}

	return b.String()
}

func registerBlockState(s *GeneralBlock) {
	h := StateHash{name: s.Name, properties: hashProperties(s.Properties)}
	if _, ok := stateRuntimeIDs[h]; ok {
		Blocks = append(Blocks, s)
		// return
		// UNSAFE !!! IGNORING SAME RUNTIME IDS !!!
		// =
		panic(fmt.Sprintf("cannot register the same state twice (%+v)", s))
	}
	rid := uint32(len(Blocks))
	if s.Name == "minecraft:air" {
		AirRID = rid
	}
	stateRuntimeIDs[h] = rid
	Blocks = append(Blocks, s)
}

type MappingIn struct {
	RIDToMCBlock       []*GeneralBlock
	NEMCRidToMCRid     []uint32
	MCRidToNEMCRid     []uint32
	NEMCRidToVal       []uint8
	NEMCToName         []string
	JavaToRid          map[string]uint32
	AirRID, NEMCAirRID uint32
}

var SchematicBlockNames = []string{
	"air",
	"stone",
	"grass",
	"dirt",
	"cobblestone",
	"planks",
	"sapling",
	"bedrock",
	"flowing_water",
	"water",
	"flowing_lava",
	"lava",
	"sand",
	"gravel",
	"gold_ore",
	"iron_ore",
	"coal_ore",
	"log",
	"leaves",
	"sponge",
	"glass",
	"lapis_ore",
	"lapis_block",
	"dispenser",
	"sandstone",
	"noteblock",
	"bed",
	"golden_rail",
	"detector_rail",
	"sticky_piston",
	"web",
	"tallgrass",
	"deadbush",
	"piston",
	"air",
	"wool",
	"air",
	"yellow_flower",
	"red_flower",
	"brown_mushroom",
	"red_mushroom",
	"gold_block",
	"iron_block",
	"double_stone_slab",
	"stone_slab",
	"brick_block",
	"tnt",
	"bookshelf",
	"mossy_cobblestone",
	"obsidian",
	"torch",
	"fire",
	"mob_spawner",
	"oak_stairs",
	"chest",
	"redstone_wire",
	"diamond_ore",
	"diamond_block",
	"crafting_table",
	"wheat",
	"farmland",
	"furnace",
	"lit_furnace",
	"standing_sign",
	"wooden_door",
	"ladder",
	"rail",
	"stone_stairs",
	"wall_sign",
	"lever",
	"stone_pressure_plate",
	"iron_door",
	"wooden_pressure_plate",
	"redstone_ore",
	"lit_redstone_ore",
	"unlit_redstone_torch",
	"redstone_torch",
	"stone_button",
	"snow_layer",
	"ice",
	"snow",
	"cactus",
	"clay",
	"reeds",
	"jukebox",
	"fence",
	"pumpkin",
	"netherrack",
	"soul_sand",
	"glowstone",
	"portal",
	"lit_pumpkin",
	"cake",
	"unpowered_repeater",
	"powered_repeater",
	"stained_glass",
	"trapdoor",
	"monster_egg",
	"stonebrick",
	"brown_mushroom_block",
	"red_mushroom_block",
	"iron_bars",
	"glass_pane",
	"melon_block",
	"pumpkin_stem",
	"melon_stem",
	"vine",
	"fence_gate",
	"brick_stairs",
	"stone_brick_stairs",
	"mycelium",
	"waterlily",
	"nether_brick",
	"nether_brick_fence",
	"nether_brick_stairs",
	"nether_wart",
	"enchanting_table",
	"brewing_stand",
	"cauldron",
	"end_portal",
	"end_portal_frame",
	"end_stone",
	"dragon_egg",
	"redstone_lamp",
	"lit_redstone_lamp",
	"double_wooden_slab",
	"wooden_slab",
	"cocoa",
	"sandstone_stairs",
	"emerald_ore",
	"ender_chest",
	"tripwire_hook",
	"tripWire",
	"emerald_block",
	"spruce_stairs",
	"birch_stairs",
	"jungle_stairs",
	"command_block",
	"beacon",
	"cobblestone_wall",
	"flower_pot",
	"carrots",
	"potatoes",
	"wooden_button",
	"skull",
	"anvil",
	"trapped_chest",
	"light_weighted_pressure_plate",
	"heavy_weighted_pressure_plate",
	"unpowered_comparator",
	"powered_comparator",
	"daylight_detector",
	"redstone_block",
	"quartz_ore",
	"hopper",
	"quartz_block",
	"quartz_stairs",
	"activator_rail",
	"dropper",
	"stained_hardened_clay",
	"stained_glass_pane",
	"leaves2",
	"log2",
	"acacia_stairs",
	"dark_oak_stairs",
	"slime",
	"barrier",
	"iron_trapdoor",
	"prismarine",
	"seaLantern",
	"hay_block",
	"carpet",
	"hardened_clay",
	"coal_block",
	"packed_ice",
	"double_plant",
	"standing_banner",
	"wall_banner",
	"daylight_detector_inverted",
	"red_sandstone",
	"red_sandstone_stairs",
	"double_stone_slab2",
	"stone_slab2",
	"spruce_fence_gate",
	"birch_fence_gate",
	"jungle_fence_gate",
	"dark_oak_fence_gate",
	"acacia_fence_gate",
	"fence",
	"fence",
	"fence",
	"fence",
	"acacia_fence_gate",
	"spruce_door",
	"birch_door",
	"jungle_door",
	"acacia_door",
	"dark_oak_door",
	"end_rod",
	"chorus_plant",
	"chorus_flower",
	"purpur_block",
	"purpur_pillar",
	"purpur_stairs",
	"purpur_double_slab",
	"purpur_slab",
	"end_bricks",
	"beetroots",
	"grass_path",
	"end_gateway",
	"repeating_command_block",
	"chain_command_block",
	"frosted_ice",
	"magma",
	"nether_wart_block",
	"red_nether_brick",
	"bone_block",
	"structure_void",
	"observer",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"shulker_box",
	"white_glazed_terracotta",
	"orange_glazed_terracotta",
	"magenta_glazed_terracotta",
	"light_blue_glazed_terracotta",
	"yellow_glazed_terracotta",
	"lime_glazed_terracotta",
	"pink_glazed_terracotta",
	"gray_glazed_terracotta",
	"silver_glazed_terracotta",
	"cyan_glazed_terracotta",
	"purple_glazed_terracotta",
	"blue_glazed_terracotta",
	"brown_glazed_terracotta",
	"green_glazed_terracotta",
	"red_glazed_terracotta",
	"black_glazed_terracotta",
	"concrete",
	"concretePowder",
	"null",
	"null",
	"structure_block",
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

func InitMapping(mappingInData []byte) {
	uncompressor := brotli.NewReader(bytes.NewBuffer(mappingInData))
	mappingIn := MappingIn{}
	if err := gob.NewDecoder(uncompressor).Decode(&mappingIn); err != nil {
		panic(err)
	}
	for _, block := range mappingIn.RIDToMCBlock {
		registerBlockState(block)
	}
	if len(Blocks) == 0 {
		panic("blockStateData read fail")
	}

	AirRID = mappingIn.AirRID
	NEMCAirRID = mappingIn.NEMCAirRID

	RuntimeIDToState = func(runtimeID uint32) (name string, properties map[string]interface{}, found bool) {
		if runtimeID >= uint32(len(Blocks)) {
			return "minecraft:air", nil, false
		}
		name, properties = Blocks[runtimeID].EncodeBlock()
		return name, properties, true
	}
	RuntimeIDToBlock = func(runtimeID uint32) (block *GeneralBlock, found bool) {
		if runtimeID >= uint32(len(Blocks)) {
			return Blocks[AirRID], false
		}
		return Blocks[runtimeID], true
	}

	StateToRuntimeID = func(name string, properties map[string]interface{}) (runtimeID uint32, found bool) {
		if rid, ok := stateRuntimeIDs[StateHash{name: name, properties: hashProperties(properties)}]; ok {
			return rid, true
		} else {
			return AirRID, false
		}
	}

	nemcToMCRIDMapping := mappingIn.NEMCRidToMCRid

	NEMCRuntimeIDToStandardRuntimeID = func(nemcRuntimeID uint32) (runtimeID uint32) {
		return uint32(nemcToMCRIDMapping[nemcRuntimeID])
	}
	StandardRuntimeIDToNEMCRuntimeID = func(runtimeID uint32) (nemcRuntimeID uint32) {
		return uint32(mappingIn.MCRidToNEMCRid[runtimeID])
	}
	if NEMCRuntimeIDToStandardRuntimeID(NEMCAirRID) != AirRID {
		panic(fmt.Errorf("air rid not matching: %d vs %d", NEMCRuntimeIDToStandardRuntimeID(NEMCAirRID), AirRID))
	}

	nemcToVal := mappingIn.NEMCRidToVal
	nemcToName := mappingIn.NEMCToName
	LegacyBlocks = make([]*LegacyBlock, len(Blocks))
	// for rid, _ := range Blocks {
	// 	LegacyBlocks[rid] = &LegacyBlock{Name: "", Val: 0}
	// }
	for nemcRid, Rid := range nemcToMCRIDMapping {
		if LegacyBlocks[Rid] != nil && LegacyBlocks[Rid].Name != "" {
			continue
		}
		val := nemcToVal[nemcRid]
		if nemcRid == int(NEMCAirRID) {
			continue
		}
		LegacyBlocks[Rid] = &LegacyBlock{Name: nemcToName[nemcRid], Val: uint16(val)}
	}
	// for rid, _ := range Blocks {
	// 	if LegacyBlocks[rid].Name == "" {
	// 		LegacyBlocks[rid].Name = "air"
	// 	}
	// }
	LegacyBlocks[AirRID].Name = "air"
	LegacyBlocks[AirRID].Val = 0
	LegacyAirBlock = LegacyBlocks[AirRID]
	for rid, block := range LegacyBlocks {
		if block != nil {
			LegacyRuntimeIDs[LegacyBlockHash{name: block.Name, data: block.Val}] = uint32(rid)
		}
	}
	LegacyRuntimeIDs[LegacyBlockHash{name: "air", data: 0}] = AirRID
	LegacyBlockToRuntimeID = func(name string, data uint16) (runtimeID uint32, found bool) {
		if rtid, hasK := LegacyRuntimeIDs[LegacyBlockHash{name: name, data: data}]; !hasK {
			return AirRID, false
		} else {
			return rtid, true
		}
	}

	{
		SchematicBlockToRuntimeIDStaticMapping = make([]uint32, 256*256)
		for _, blkName := range SchematicBlockNames {
			if blkName == "null" {
				continue
			}
			if _, found := LegacyBlockToRuntimeID(blkName, 0); !found {
				// TODO: fix schematic block mapping
				//fmt.Printf("Warning schematic block %v not found\n", blkName)
			}
		}
		notFound := ^uint32(0)
		var name string
		for block := byte(0); block < 255; block++ {
			for data := byte(0); data < 255; data++ {
				index := uint16(block)<<8 | uint16(data)
				name = SchematicBlockNames[block]
				rtidU32, found := LegacyBlockToRuntimeID(name, (uint16(data)))
				if !found {
					rtidU32, found = LegacyBlockToRuntimeID(name, 0)
					if !found {
						rtidU32 = notFound
					}
				}
				SchematicBlockToRuntimeIDStaticMapping[index] = rtidU32
			}
		}
		SchematicBlockToRuntimeID = func(block, data byte) (runtimeID uint32, found bool) {
			index := uint16(block)<<8 | uint16(data)
			if rtid := SchematicBlockToRuntimeIDStaticMapping[index]; rtid == notFound {
				return AirRID, false
			} else {
				return rtid, true
			}
		}
	}

	RuntimeIDToLegacyBlock = func(runtimeID uint32) (legacyBlock *LegacyBlock, found bool) {
		if blk := LegacyBlocks[runtimeID]; blk == nil {
			return LegacyAirBlock, false
		} else {
			return blk, true
		}
	}
	numberRegex := regexp.MustCompile(`\d+`)
	legacyBlockNameRegex := regexp.MustCompile(`name=.+,`)
	legacyBlockValRegex := regexp.MustCompile(`val=.+]`)
	JavaStrToRuntimeIDMapping = mappingIn.JavaToRid
	JavaStrPropsToRuntimeIDMapping = make(map[string]map[string]uint32)
	splitJavaNameAndProps := func(javaName string) (name, prop string) {
		ss := strings.Split(javaName, "[")
		if len(ss) == 1 {
			return ss[0], ""
		}
		return ss[0], strings.TrimRight(ss[1], "]")
	}
	{
		for javaName, rtid := range JavaStrToRuntimeIDMapping {
			jname, jprop := splitJavaNameAndProps(javaName)
			if props, found := JavaStrPropsToRuntimeIDMapping[jname]; !found {
				JavaStrPropsToRuntimeIDMapping[jname] = map[string]uint32{jprop: rtid}
			} else {
				props[jprop] = rtid
			}
		}

	}
	JavaToRuntimeID = func(javaBlockStr string) (runtimeID uint32, found bool) {
		if rtid, hasK := JavaStrToRuntimeIDMapping[javaBlockStr]; hasK {
			return rtid, true
		} else if strings.HasPrefix(javaBlockStr, "omega:as_runtime_id[") {
			matchs := numberRegex.FindAllString(javaBlockStr, 1)
			if len(matchs) > 0 {
				if rtid, err := strconv.Atoi(string(matchs[0])); err == nil {
					mappingIn.JavaToRid[javaBlockStr] = uint32(rtid)
					return uint32(rtid), true
				}
			}
			mappingIn.JavaToRid[javaBlockStr] = AirRID
		}
		jname, jprop := splitJavaNameAndProps(javaBlockStr)
		if oprops, found := JavaStrPropsToRuntimeIDMapping[jname]; found {
			bscore := -1
			brtid := AirRID
			tprop := strings.Split(jprop, ",")
			for prop, rtid := range oprops {
				oprop := strings.Split(prop, ",")
				score := 0
				for _, p := range tprop {
					for _, m := range oprop {
						if m == p {
							score++
							break
						}
					}
				}
				if score > bscore {
					bscore = score
					brtid = rtid
				}
			}
			return brtid, true
		}
		if strings.HasPrefix(javaBlockStr, "omega:as_legacy_block[") {
			name := "air"
			matchs := legacyBlockNameRegex.FindAllString(javaBlockStr, 1)
			if len(matchs) > 0 {
				name = matchs[0]
				name = name[5 : len(name)-1]
			}
			matchs = legacyBlockValRegex.FindAllString(javaBlockStr, 1)
			if len(matchs) > 0 {
				if val, err := strconv.Atoi(string(matchs[0][4 : len(matchs[0])-1])); err == nil {
					rtid, found := LegacyBlockToRuntimeID(name, uint16(val))
					if found {
						mappingIn.JavaToRid[javaBlockStr] = uint32(rtid)
						return uint32(rtid), true
					}

				}
			}
			mappingIn.JavaToRid[javaBlockStr] = AirRID
		}
		return AirRID, false
	}
	RuntimeIDToJavaStrMapping = make(map[uint32]string)
	for javaStr, rtid := range mappingIn.JavaToRid {
		RuntimeIDToJavaStrMapping[rtid] = javaStr
	}
	RuntimeIDToJava = func(runtimeID uint32) (javaBlockStr string, found bool) {
		if javaBlockStr, hasK := RuntimeIDToJavaStrMapping[runtimeID]; hasK {
			return javaBlockStr, true
		} else {
			return JavaAirBlock, false
		}
	}

}

//go:embed blockmapping_nemc_2_5_15_mc_1_19.gob.brotli
var mappingInData []byte

func init() {
	InitMapping(mappingInData)
}
