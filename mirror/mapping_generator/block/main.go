package main

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"

	"github.com/andybalholm/brotli"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

//go:embed block_states_1_18_30.nbt
var blockStateData []byte

//go:embed runtimeIds_2_1_10.json
var nemcJsonData []byte

type GeneralBlock struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

type RichBlock struct {
	GeneralBlock
	Val     int
	NEMCRID int
	RID     int
}

type NEMCBlock struct {
	Name    string
	Val     int
	NEMCRID int
}

func NewNEMCBlock(p [2]interface{}, nemcRID int) NEMCBlock {
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
	runtimeIDData := make([][2]interface{}, 0)
	err := json.Unmarshal(nemcJsonData, &runtimeIDData)
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
	RIDToMCBlock   []*GeneralBlock
	NEMCRidToMCRid []int16
	NEMCRidToVal   []uint8
}

func main() {
	airRID := 0
	nemcAirRID := 0
	skullRID := 0
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
			NEMCRID:      -1,
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
		if skullRID == 0 && rb.Name == "minecraft:skull" {
			skullRID = rid
			rb.Val = 0
		}
		blocks = append(blocks, rb)
		ridToMCBlock = append(ridToMCBlock, &GeneralBlock{
			Name:       s.Name,
			Properties: s.Properties,
			Version:    s.Version,
		})
	}

	if err := os.MkdirAll("convert_out", 0755); err != nil {
		panic(err)
	}
	fp, err := os.OpenFile("convert_out/StandardMCStates.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(fp)
	enc.SetIndent("", "\t")
	enc.Encode(groupedBlocks)
	fp.Close()
	nemcData := ReadNemcData()
	for rid, nemcBlocks := range nemcData {
		if nemcBlocks.Name == "air" {
			nemcAirRID = rid
		}
		if group, found := groupedBlocks[Remapping(nemcBlocks.Name)]; !found {
			fmt.Println("not found: ", nemcBlocks)
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
				fmt.Println("not found: ", nemcBlocks)
			}
		}
	}
	fmt.Println(airRID, " ", nemcAirRID, " ", skullRID)

	nemcToMCRIDMapping := make([]int16, len(nemcData))
	nemcToVal := make([]uint8, len(nemcData))
	for i := 0; i < len(nemcData); i++ {
		nemcToMCRIDMapping[i] = int16(airRID)
		nemcToVal[i] = uint8(nemcData[i].Val)
	}
	for rid, b := range blocks {
		if b.NEMCRID == -1 {
			fmt.Println("new block ", b)
			continue
		}
		origBlock := nemcData[b.NEMCRID]
		if origBlock.Name == "minecraft:skull" {
			nemcToMCRIDMapping[b.NEMCRID] = int16(skullRID)
			continue
		}
		nemcToMCRIDMapping[b.NEMCRID] = int16(rid)
	}
	fmt.Println(nemcToMCRIDMapping[nemcAirRID])
	fp, err = os.OpenFile("convert_out/nemcRIDToMC1_18_30RID.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err = json.NewEncoder(fp).Encode(nemcToMCRIDMapping); err != nil {
		panic(err)
	}
	fp.Close()

	fp, err = os.OpenFile("convert_out/nemcRIDToVal.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err = json.NewEncoder(fp).Encode(nemcToVal); err != nil {
		panic(err)
	}
	fp.Close()
	mapping_out := MappingOut{
		RIDToMCBlock:   ridToMCBlock,
		NEMCRidToMCRid: nemcToMCRIDMapping,
		NEMCRidToVal:   nemcToVal,
	}
	fp, err = os.OpenFile("convert_out/blockmapping_nemc_1_17_0_mc_1_18_30.gob", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	compressor := brotli.NewWriter(fp)
	if err := gob.NewEncoder(compressor).Encode(mapping_out); err != nil {
		panic(err)
	}
	if err := compressor.Close(); err != nil {
		panic(err)
	}
	fp.Close()
}
