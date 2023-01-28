package main

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"phoenixbuilder/minecraft/nbt"

	"github.com/andybalholm/brotli"
)

//go:embed block_states_1_19.nbt
var blockStateData []byte

//go:embed block_1_18_java_to_bedrock.json
var javaJsonData []byte

type GeneralBlock struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

type RichBlock struct {
	GeneralBlock
	Val        int
	NEMCRID    int
	RID        int
	JavaString string
}

type NEMCBlock struct {
	Name    string
	Val     int
	NEMCRID int
}

type JavaToBedrockMappingIn struct {
	Name       string         `json:"bedrock_identifier"`
	Properties map[string]any `json:"bedrock_states"`
}

type IDGroup struct {
	Count int
	IDS   []*RichBlock
}

func NewIDGroup() *IDGroup {
	return &IDGroup{
		Count: 0,
		IDS:   make([]*RichBlock, 0),
	}
}

func (ig *IDGroup) AppendItem(p *RichBlock) {
	p.Val = len(ig.IDS)
	ig.IDS = append(ig.IDS, p)
}

func ReadNemcData() []NEMCBlock {
	var nemcJsonData []byte
	nemcJsonData, err := ioutil.ReadFile("resources/blockRuntimeIDs/netease/runtimeIds_2_5_15_proc.json")
	if err != nil {
		panic(err)
	}

	NewNEMCBlock := func(p [2]interface{}, nemcRID int) NEMCBlock {
		s, ok := p[0].(string)
		if !ok {
			panic("fail")
		}
		i, ok := p[1].(float64)
		if !ok {
			panic("fail")
		}
		return NEMCBlock{s, int(i), nemcRID}
	}

	runtimeIDData := make([][2]interface{}, 0)
	err = json.Unmarshal(nemcJsonData, &runtimeIDData)
	if err != nil {
		panic(err)
	}
	nemcBlocks := make([]NEMCBlock, 0)
	for rid, jd := range runtimeIDData {
		nemcBlocks = append(nemcBlocks, NewNEMCBlock(jd, rid))
	}
	return nemcBlocks
}

var remapper = map[string]string{
	"concretePowder":           "concrete_powder",
	"invisibleBedrock":         "invisible_bedrock",
	"movingBlock":              "moving_block",
	"pistonArmCollision":       "piston_arm_collision",
	"stickyPistonArmCollision": "sticky_piston_arm_collision",
	"tripWire":                 "trip_wire",
	"seaLantern":               "sea_lantern",
}

func Remapping(nemcName string) (mcName string) {
	if newN, found := remapper[nemcName]; found {
		return "minecraft:" + newN
	}
	return "minecraft:" + nemcName
}

type MappingOut struct {
	RIDToMCBlock       []*GeneralBlock
	NEMCRidToMCRid     []uint32
	MCRidToNEMCRid     []uint32
	NEMCRidToVal       []uint8
	NEMCToName         []string
	JavaToRid          map[string]uint32
	AirRID, NEMCAirRID uint32
}

func main() {
	for i := 1; i < 5; i++ {
		if i == 1 {
			remapper["stone_slab"] = "stone_block_slab"
			remapper["double_stone_slab"] = "double_stone_block_slab"
		} else {
			remapper[fmt.Sprintf("stone_slab%v", i)] = fmt.Sprintf("stone_block_slab%v", i)
			remapper[fmt.Sprintf("double_stone_slab%v", i)] = fmt.Sprintf("double_stone_block_slab%v", i)
		}
	}
	for _, color := range []string{"purple", "pink", "green", "red", "gray", "light_blue", "yellow", "blue", "brown", "black", "white", "orange", "cyan", "magenta", "lime", "silver"} {
		// "glazedTerracotta.purple":  "purple_glazed_terracotta",
		remapper[fmt.Sprintf("glazedTerracotta.%v", color)] = fmt.Sprintf("%v_glazed_terracotta", color)
	}

	if err := os.MkdirAll("convert_out", 0755); err != nil {
		panic(err)
	}

	// group up standard mc block states
	airRID := 0
	nemcAirRID := 0
	NEMCRIDNOTFOUND := -1
	dec := nbt.NewDecoder(bytes.NewBuffer(blockStateData))
	groupedBlocks := make(map[string]*IDGroup)
	blocks := []*RichBlock{}
	ridToMCBlock := []*GeneralBlock{}
	for rid := 0; ; rid++ {
		var s GeneralBlock
		if err := dec.Decode(&s); err != nil {
			break
		}
		rb := &RichBlock{
			GeneralBlock: s,
			Val:          0,
			NEMCRID:      NEMCRIDNOTFOUND,
			RID:          int(rid),
		}
		_, hasK := groupedBlocks[rb.Name]
		if rb.Name == "minecraft:air" {
			airRID = rid
		}

		if !hasK {
			groupedBlocks[rb.Name] = NewIDGroup()
		}
		groupedBlocks[rb.Name].AppendItem(rb)
		blocks = append(blocks, rb)
		ridToMCBlock = append(ridToMCBlock, &GeneralBlock{
			Name:       s.Name,
			Properties: s.Properties,
			Version:    s.Version,
		})
	}
	var fp *os.File
	var err error

	// read nemc blocks and generated mapping
	nemcData := ReadNemcData()
	for rid, nemcBlocks := range nemcData {
		if nemcBlocks.Name == "air" {
			nemcAirRID = rid
		}
		if group, found := groupedBlocks[Remapping(nemcBlocks.Name)]; !found {
			fmt.Printf("Nemc block: %v not found in MC.\n", nemcBlocks)
		} else {
			found := false
			for val, b := range group.IDS {
				if b.NEMCRID == -1 && b.Val == val {
					b.NEMCRID = rid
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("Nemc block-(value): %v not found in MC.\n", nemcBlocks)
			}
		}
	}
	fmt.Println("MC Air RID:   ", airRID)
	fmt.Println("NEMC Air RID: ", nemcAirRID)

	nemcToMCRIDMapping := make([]uint32, len(nemcData))

	nemcToVal := make([]uint8, len(nemcData))
	nemcToName := make([]string, len(nemcData))
	for i := 0; i < len(nemcData); i++ {
		nemcToMCRIDMapping[i] = uint32(airRID)
		nemcToVal[i] = uint8(nemcData[i].Val)
		nemcToName[i] = nemcData[i].Name
	}
	mcRtidToNEMCRid := make([]uint32, len(blocks))
	for rid, b := range blocks {
		if b.NEMCRID == NEMCRIDNOTFOUND {
			fmt.Printf("MC block %v not found in NEMC\n", b)
			mcRtidToNEMCRid[rid] = uint32(nemcAirRID)
			continue
		}
		mcRtidToNEMCRid[rid] = uint32(b.NEMCRID)
		// origBlock := nemcData[b.NEMCRID]
		// if origBlock.Name == "minecraft:skull" {
		// 	nemcToMCRIDMapping[b.NEMCRID] = int16(skullRID)
		// 	continue
		// }
		nemcToMCRIDMapping[b.NEMCRID] = uint32(rid)
	}
	fmt.Println(nemcToMCRIDMapping[nemcAirRID])
	fp, err = os.OpenFile("convert_out/nemcRIDToMC1_19RID.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(fp).Encode(nemcToMCRIDMapping); err != nil {
		panic(err)
	}
	fp.Close()

	fp, err = os.OpenFile("convert_out/nemcRIDToVal.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(fp).Encode(nemcToVal); err != nil {
		panic(err)
	}
	fp.Close()

	// java mapping
	javaBlocks := map[string]JavaToBedrockMappingIn{}
	err = json.Unmarshal(javaJsonData, &javaBlocks)
	if err != nil {
		panic(err)
	}
	javaToRid := map[string]uint32{}
	for javaName, bedrockBlockDescribe := range javaBlocks {
		// if javaName == "minecraft:campfire[facing=east,lit=false,signal_fire=false,waterlogged=false]" {
		// 	fmt.Println("stop")
		// }
		bedrockBlocks, hasK := groupedBlocks[bedrockBlockDescribe.Name]
		if !hasK {
			fmt.Println(javaName, " group not found")
			continue
		}
		rbs := bedrockBlocks.IDS
		found := false
		for _, rb := range rbs {
			propTarget := rb.Properties
			propSrc := bedrockBlockDescribe.Properties
			propMatch := true
			for k, targetV := range propTarget {
				wanT := reflect.TypeOf(targetV).Kind()
				srcV := propSrc[k]
				switch wanT {
				case reflect.Uint8:
					if b, ok := srcV.(bool); ok {
						if b != (targetV.(uint8) == 1) {
							propMatch = false
						}
					} else {
						if uint8(int(srcV.(float64))) != targetV.(uint8) {
							propMatch = false
						}
					}
				case reflect.Int32:
					if int32(int(srcV.(float64))) != targetV.(int32) {
						propMatch = false
					}
				case reflect.String:
					if srcV.(string) != targetV.(string) {
						propMatch = false
					}
				default:
					panic(wanT)
				}
				if !propMatch {
					break
				}
			}
			if propMatch {
				found = true
				rb.JavaString = javaName
				javaToRid[javaName] = uint32(rb.RID)
				javaNameAlter := strings.ReplaceAll(javaName, ",waterlogged=false", "")
				if javaNameAlter != javaName {
					javaToRid[javaNameAlter] = uint32(rb.RID)
				}
				javaNameAlter = strings.ReplaceAll(javaName, ",waterlogged=true", "")
				if javaNameAlter != javaName {
					javaToRid[javaNameAlter] = uint32(rb.RID)
				}
				javaNameAlter = strings.ReplaceAll(javaName, "waterlogged=false,", "")
				if javaNameAlter != javaName {
					javaToRid[javaNameAlter] = uint32(rb.RID)
				}
				javaNameAlter = strings.ReplaceAll(javaName, "waterlogged=true,", "")
				if javaNameAlter != javaName {
					javaToRid[javaNameAlter] = uint32(rb.RID)
				}
				javaNameAlter = strings.ReplaceAll(javaName, "waterlogged=true", "")
				if javaNameAlter != javaName {
					javaToRid[javaNameAlter] = uint32(rb.RID)
				}
				break
			}
		}
		if !found {
			fmt.Println(javaName, " block statue not found")
		}
	}

	counter := 0
	for _, blk := range blocks {
		if blk.JavaString == "" {
			counter += 1
			// fmt.Printf("%v not found in java\n", blk)
		}
	}
	fmt.Printf("%v blocks not found in java\n", counter)

	fp, err = os.OpenFile("convert_out/javaBlockToRid.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(fp).Encode(javaToRid); err != nil {
		panic(err)
	}
	fp.Close()

	mapping_out := MappingOut{
		RIDToMCBlock:   ridToMCBlock,
		NEMCRidToMCRid: nemcToMCRIDMapping,
		MCRidToNEMCRid: mcRtidToNEMCRid,
		NEMCRidToVal:   nemcToVal,
		NEMCToName:     nemcToName,
		JavaToRid:      javaToRid,
		AirRID:         uint32(airRID),
		NEMCAirRID:     uint32(nemcAirRID),
	}

	fp, err = os.OpenFile("convert_out/StandardMCStates.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(fp)
	enc.SetIndent("", "\t")
	enc.Encode(groupedBlocks)
	fp.Close()

	fp, err = os.OpenFile("mirror/chunk/blockmapping_nemc_2_5_15_mc_1_19.gob.brotli", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	compressor := brotli.NewWriter(fp)
	if err := gob.NewEncoder(compressor).Encode(mapping_out); err != nil {
		panic(err)
	}
	if err := compressor.Close(); err != nil {
		panic(err)
	}
	fp.Close()
}
