package chunk

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"fmt"
	"sort"
	"strings"
	"unsafe"

	"github.com/andybalholm/brotli"
)

var StateToRuntimeID func(name string, properties map[string]any) (runtimeID uint32, found bool)
var RuntimeIDToState func(runtimeID uint32) (name string, properties map[string]any, found bool)
var RuntimeIDToBlock func(runtimeID uint32) (block *GeneralBlock, found bool)
var NEMCRuntimeIDToStandardRuntimeID func(nemcRuntimeID uint32) (runtimeID uint32)
var RuntimeIDToLegacyBlock func(runtimeID uint32) (legacyBlock *LegacyBlock)
var LegacyBlockToRuntimeID func(name string, data uint8) (runtimeID uint32, found bool)
var AirRID uint32
var NEMCAirRID uint32
var stateRuntimeIDs = map[StateHash]uint32{}
var Blocks []*GeneralBlock
var LegacyBlocks []*LegacyBlock
var LegacyRuntimeIDs = map[LegacyBlockHash]uint32{}

var JavaToRuntimeID func(javaBlockStr string) (runtimeID uint32, found bool)
var JavaStrToRuntimeIDMapping map[string]uint32

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
	Val  byte
}

func (b GeneralBlock) EncodeBlock() (string, map[string]any) {
	return b.Name, b.Properties
}

type StateHash struct {
	name, properties string
}

type LegacyBlockHash struct {
	name string
	data uint8
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
	RIDToMCBlock   []*GeneralBlock
	NEMCRidToMCRid []int16
	NEMCRidToVal   []uint8
	NEMCToName     []string
	JavaToRid      map[string]uint32
}

var SchematicBlockMapping = []string{
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

	RuntimeIDToState = func(runtimeID uint32) (name string, properties map[string]interface{}, found bool) {
		if runtimeID >= uint32(len(Blocks)) {
			return "", nil, false
		}
		name, properties = Blocks[runtimeID].EncodeBlock()
		return name, properties, true
	}
	RuntimeIDToBlock = func(runtimeID uint32) (block *GeneralBlock, found bool) {
		if runtimeID >= uint32(len(Blocks)) {
			return nil, false
		}
		return Blocks[runtimeID], true
	}

	StateToRuntimeID = func(name string, properties map[string]interface{}) (runtimeID uint32, found bool) {
		rid, ok := stateRuntimeIDs[StateHash{name: name, properties: hashProperties(properties)}]
		return rid, ok
	}

	nemcToMCRIDMapping := mappingIn.NEMCRidToMCRid

	NEMCAirRID = 134
	NEMCRuntimeIDToStandardRuntimeID = func(nemcRuntimeID uint32) (runtimeID uint32) {
		return uint32(nemcToMCRIDMapping[nemcRuntimeID])
	}
	if NEMCRuntimeIDToStandardRuntimeID(NEMCAirRID) != AirRID {
		panic(fmt.Errorf("Air rid not matching: %d vs %d.", NEMCRuntimeIDToStandardRuntimeID(NEMCAirRID), AirRID))
	}

	nemcToVal := mappingIn.NEMCRidToVal
	nemcToName := mappingIn.NEMCToName
	LegacyBlocks = make([]*LegacyBlock, len(Blocks))
	for rid, _ := range Blocks {
		LegacyBlocks[rid] = &LegacyBlock{Name: "", Val: 0}
	}
	for nemcRid, Rid := range nemcToMCRIDMapping {
		if LegacyBlocks[Rid].Name != "" {
			continue
		}
		val := nemcToVal[nemcRid]
		if nemcRid == int(NEMCAirRID) {
			continue
		}
		LegacyBlocks[Rid].Val = val
		LegacyBlocks[Rid].Name = nemcToName[nemcRid]
	}
	for rid, _ := range Blocks {
		if LegacyBlocks[rid].Name == "" {
			LegacyBlocks[rid].Name = "air"
		}
	}
	LegacyBlocks[AirRID].Name = "air"
	LegacyBlocks[AirRID].Val = 0
	for rid, block := range LegacyBlocks {
		LegacyRuntimeIDs[LegacyBlockHash{name: block.Name, data: block.Val}] = uint32(rid)
	}
	LegacyRuntimeIDs[LegacyBlockHash{name: "air", data: 0}] = AirRID
	LegacyBlockToRuntimeID = func(name string, data uint8) (runtimeID uint32, found bool) {
		if rtid, hasK := LegacyRuntimeIDs[LegacyBlockHash{name: name, data: data}]; !hasK {
			return AirRID, false
		} else {
			return rtid, true
		}
	}
	for _, blkName := range SchematicBlockMapping {
		if blkName == "null" {
			continue
		}
		if _, found := LegacyBlockToRuntimeID(blkName, 0); !found {
			//fmt.Printf("Warning schematic block %v not found\n", blkName)
		}
	}
	RuntimeIDToLegacyBlock = func(runtimeID uint32) (legacyBlock *LegacyBlock) {
		return LegacyBlocks[runtimeID]
	}
	JavaStrToRuntimeIDMapping = mappingIn.JavaToRid
	JavaToRuntimeID = func(javaBlockStr string) (runtimeID uint32, found bool) {
		if rtid, hasK := mappingIn.JavaToRid[javaBlockStr]; hasK {
			return rtid, true
		} else {
			return AirRID, false
		}
	}
}

//go:embed blockmapping_nemc_2_2_15_mc_1_19.gob.brotli
var mappingInData []byte

func init() {
	InitMapping(mappingInData)
}
