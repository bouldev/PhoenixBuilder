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
var StatePropsToRuntimeIDMapping map[string]map[string]uint32
var RuntimeIDToState func(runtimeID uint32) (name string, properties map[string]any, found bool)
var RuntimeIDToStateStr func(runtimeID uint32) (blockNameWithState string, found bool)
var RuntimeIDToBlock func(runtimeID uint32) (block *GeneralBlock, found bool)
var JavaToRuntimeID func(javaBlockStr string) (runtimeID uint32, found bool)
var RuntimeIDToJava func(runtimeID uint32) (javaBlockStr string, found bool)
var BlockStateStrToRuntimeID func(blockName, blockState string) (uint32, bool)
var BlockPropsToRuntimeID func(blockName string, blockProps map[string]interface{}) (uint32, bool)

var SchematicBlockToRuntimeID func(block, data byte) (runtimeID uint32, found bool)
var SchematicBlockToRuntimeIDStaticMapping []uint32
var stateRuntimeIDs = map[StateHash]uint32{}
var Blocks []*GeneralBlock
var LegacyBlocks []*LegacyBlock
var LegacyRuntimeIDs = map[LegacyBlockHash]uint32{}
var JavaStrToRuntimeIDMapping map[string]uint32
var JavaStrPropsToRuntimeIDMapping map[string]map[string]uint32
var RuntimeIDToJavaStrMapping map[uint32]string
var RuntimeIDToSateStrMapping map[uint32]string

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

func PropValueToStateString(value interface{}) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		} else {
			return "false"
		}
	case uint8:
		if v == 0 {
			return "false"
		} else if v == 1 {
			return "true"
		} else {
			return fmt.Sprintf("%v", v)
		}
	case int32:
		return fmt.Sprintf("%v", v)
	case string:
		return fmt.Sprintf("\"%v\"", v)
	default:
		// If block encoding is broken, we want to find out as soon as possible. This saves a lot of time
		// debugging in-game.
		return fmt.Sprintf("%v", v)
	}
}

func PropsToStateString(properties map[string]interface{}, bracket bool) string {
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

	props := make([]string, 0, len(properties))
	for _, k := range keys {
		switch v := properties[k].(type) {
		case bool:
			if v {
				props = append(props, fmt.Sprintf("\"%v\": true", k))
			} else {
				props = append(props, fmt.Sprintf("\"%v\": false", k))
			}
		case uint8:
			if v == 0 {
				props = append(props, fmt.Sprintf("\"%v\": false", k))
			} else if v == 1 {
				props = append(props, fmt.Sprintf("\"%v\": true", k))
			} else {
				props = append(props, fmt.Sprintf("\"%v\": %v", k, v))
			}
		case int32:
			props = append(props, fmt.Sprintf("\"%v\": %v", k, v))
		case string:
			props = append(props, fmt.Sprintf("\"%v\": \"%v\"", k, v))
		default:
			// If block encoding is broken, we want to find out as soon as possible. This saves a lot of time
			// debugging in-game.
			panic(fmt.Sprintf("invalid block property type %T for property %v", v, k))
		}
	}

	stateStr := strings.Join(props, ", ")
	if !bracket {
		return stateStr
	}
	return fmt.Sprintf("[%v]", stateStr)
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

	props := PropsToStateString(s.Properties, false)

	RuntimeIDToSateStrMapping[rid] = fmt.Sprintf("%v [%v]", strings.TrimPrefix(s.Name, "minecraft:"), props)

	if g, found := StatePropsToRuntimeIDMapping[s.Name]; found {
		g[props] = rid
	} else {
		StatePropsToRuntimeIDMapping[s.Name] = map[string]uint32{props: rid}
	}
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

func InitMapping(mappingInData []byte) {
	//StdToNemcBlockNameMapping := map[string]string{}
	//NemcToStdBlockNameMapping := map[string]string{}
	//{
	//	for i := 1; i < 5; i++ {
	//		if i == 1 {
	//			NemcToStdBlockNameMapping["stone_slab"] = "stone_block_slab"
	//			NemcToStdBlockNameMapping["double_stone_slab"] = "double_stone_block_slab"
	//		} else {
	//			NemcToStdBlockNameMapping[fmt.Sprintf("stone_slab%v", i)] = fmt.Sprintf("stone_block_slab%v", i)
	//			NemcToStdBlockNameMapping[fmt.Sprintf("double_stone_slab%v", i)] = fmt.Sprintf("double_stone_block_slab%v", i)
	//		}
	//	}
	//	for _, color := range []string{"purple", "pink", "green", "red", "gray", "light_blue", "yellow", "blue", "brown", "black", "white", "orange", "cyan", "magenta", "lime", "silver"} {
	//		// "glazedTerracotta.purple":  "purple_glazed_terracotta",
	//		NemcToStdBlockNameMapping[fmt.Sprintf("glazedTerracotta.%v", color)] = fmt.Sprintf("%v_glazed_terracotta", color)
	//	}
	//	for k, v := range NemcToStdBlockNameMapping {
	//		StdToNemcBlockNameMapping[v] = k
	//		StdToNemcBlockNameMapping["minecraft:"+v] = "minecraft:" + k
	//	}
	//	for k, v := range StdToNemcBlockNameMapping {
	//		NemcToStdBlockNameMapping[v] = k
	//	}
	//}

	uncompressor := brotli.NewReader(bytes.NewBuffer(mappingInData))
	mappingIn := MappingIn{}
	if err := gob.NewDecoder(uncompressor).Decode(&mappingIn); err != nil {
		panic(err)
	}
	blockNameRegrades := map[string]string{}
	for i, blk := range mappingIn.RIDToMCBlock {
		if strings.HasPrefix(blk.Name, "minecraft:double_stone_block_slab") {
			upperGradeName := blk.Name
			blk.Name = strings.ReplaceAll(blk.Name, "minecraft:double_stone_block_slab", "minecraft:double_stone_slab")
			lowerGradeName := blk.Name
			mappingIn.RIDToMCBlock[i] = blk
			blockNameRegrades[upperGradeName] = lowerGradeName
		}
		if strings.HasPrefix(blk.Name, "minecraft:stone_block_slab") {
			upperGradeName := blk.Name
			blk.Name = strings.ReplaceAll(blk.Name, "minecraft:stone_block_slab", "minecraft:stone_slab")
			lowerGradeName := blk.Name
			mappingIn.RIDToMCBlock[i] = blk
			blockNameRegrades[upperGradeName] = lowerGradeName
		}
		if strings.HasSuffix(blk.Name, "_glazed_terracotta") {
			upperGradeName := blk.Name
			lowerGradeName := strings.ReplaceAll(upperGradeName, "_glazed_terracotta", "")
			lowerGradeName = "minecraft:glazedTerracotta." + lowerGradeName[len("minecraft:"):]
			blockNameRegrades[lowerGradeName] = upperGradeName
		}
		if strings.HasSuffix(blk.Name, "sea_lantern") {
			upperGradeName := blk.Name
			lowerGradeName := strings.ReplaceAll(upperGradeName, "sea_lantern", "seaLantern")
			blockNameRegrades[lowerGradeName] = upperGradeName
		}
		if strings.HasSuffix(blk.Name, "trip_wire") {
			upperGradeName := blk.Name
			lowerGradeName := strings.ReplaceAll(upperGradeName, "trip_wire", "tripWire")
			blockNameRegrades[lowerGradeName] = upperGradeName
		}
		if strings.HasSuffix(blk.Name, "concrete_powder") {
			upperGradeName := blk.Name
			lowerGradeName := strings.ReplaceAll(upperGradeName, "concrete_powder", "concretePowder")
			blockNameRegrades[lowerGradeName] = upperGradeName
		}
	}

	blockNameRedirect := func(origBlockName string) (stdMCBlockName string) {
		origBlockName = strings.TrimSpace(origBlockName)
		if !strings.HasPrefix(origBlockName, "minecraft:") {
			origBlockName = "minecraft:" + origBlockName
		}
		if regradeName, found := blockNameRegrades[origBlockName]; found {
			return regradeName
		} else {
			return origBlockName
		}
	}

	StatePropsToRuntimeIDMapping = make(map[string]map[string]uint32)
	RuntimeIDToSateStrMapping = make(map[uint32]string)

	RuntimeIDToStateStr = func(rtid uint32) (blockNameWithState string, found bool) {
		blockNameWithState, found = RuntimeIDToSateStrMapping[rtid]
		return blockNameWithState, found
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
	JavaStrToRuntimeIDMapping = mappingIn.JavaToRid
	JavaStrPropsToRuntimeIDMapping = make(map[string]map[string]uint32)
	splitJavaNameAndProps := func(javaName string) (name, prop string) {
		ss := strings.Split(javaName, "[")
		if len(ss) == 1 {
			return ss[0], ""
		}
		return ss[0], strings.TrimRight(ss[1], "]")
	}
	trimStateProps := func(state string) (prop string) {
		state = strings.TrimSpace(state)
		return strings.TrimRight(strings.TrimLeft(state, "["), "]")
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
	BlockPropsToRuntimeID = func(blockName string, blockProps map[string]interface{}) (uint32, bool) {
		blockName = blockNameRedirect(blockName)
		if oprops, found := StatePropsToRuntimeIDMapping[blockName]; found {
			bscore := -1
			brtid := AirRID
			for prop, rtid := range oprops {
				oprop := strings.Split(prop, ",")
				score := 0
				for k, v := range blockProps {
					p := fmt.Sprintf("\"%v\":%v", k, PropValueToStateString(v))
					for _, m := range oprop {
						m = strings.ReplaceAll(m, " ", "")
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
		return AirRID, false
	}
	BlockStateStrToRuntimeID = func(blockName, blockState string) (uint32, bool) {
		blockName = blockNameRedirect(blockName)
		sprops := trimStateProps(blockState)
		if oprops, found := StatePropsToRuntimeIDMapping[blockName]; found {
			bscore := -1
			brtid := AirRID
			tprop := strings.Split(sprops, ",")
			for prop, rtid := range oprops {
				oprop := strings.Split(prop, ",")
				score := 0
				for _, p := range tprop {
					p = strings.ReplaceAll(p, " ", "")
					for _, m := range oprop {
						m = strings.ReplaceAll(m, " ", "")
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
		return AirRID, false
	}
	JavaToRuntimeID = func(javaBlockStr string) (runtimeID uint32, found bool) {
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
