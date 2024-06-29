package step2_add_specific_legacy_converts

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"phoenixbuilder/mirror/blocks/block_set"
	"phoenixbuilder/mirror/blocks/convertor"
	"phoenixbuilder/mirror/blocks/describe"
	"sort"
	"strings"
)

type RawState struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func (s RawState) ToValue() describe.PropVal {
	if s.Type == "string" {
		return describe.PropValFromString(s.Value.(string))
	} else if s.Type == "int" {
		return describe.PropValFromInt32(int32(s.Value.(float64)))
	} else if s.Type == "byte" {
		if s.Value.(float64) == 0 {
			return describe.PropVal0
		} else if s.Value.(float64) == 1 {
			return describe.PropVal1
		} else {
			panic(s.Value)
		}
	} else {
		panic(s.Type)
	}
}

type RawBlockPalette struct {
	LegacyData uint16     `json:"data"` // up to 5469 @ cobblestone_wall
	BlockName  string     `json:"name"`
	States     []RawState `json:"states"`
}

func (p RawBlockPalette) DumpStates() (StateOrder []string, State map[string]describe.PropVal, States describe.Props) {
	StateOrder = []string{}
	State = map[string]describe.PropVal{}
	States = describe.Props{}
	for _, rawState := range p.States {
		p := rawState.ToValue()
		State[rawState.Name] = p
		StateOrder = append(StateOrder, rawState.Name)

		States = append(States, struct {
			Name  string
			Value describe.PropVal
		}{
			Name:  rawState.Name,
			Value: p,
		})
	}
	if !sort.StringsAreSorted(StateOrder) {
		fmt.Println(StateOrder)
	}
	return StateOrder, State, States
}

type RawData struct {
	Blocks []RawBlockPalette `json:"blocks"`
}

type ParsedBlock struct {
	NameWithoutMC string
	States        describe.Props
	// State         map[string]blocks.PropVal
	// StateOrder    []string
	LegacyData    uint16
	NemcRuntimeID int32
	Version       int32
}

func ConvertRawData(rawData *RawData) []ParsedBlock {
	b := []byte{1, 20, 10, 0}
	Version := int32(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
	parsedBlocks := []ParsedBlock{}
	for rtid, block := range rawData.Blocks {
		_, _, States := block.DumpStates()
		parsedBlock := ParsedBlock{
			NameWithoutMC: strings.TrimPrefix(block.BlockName, "minecraft:"),
			States:        States,
			LegacyData:    block.LegacyData,
			NemcRuntimeID: int32(rtid),
			// TODO: version
			Version: Version,
		}
		parsedBlocks = append(parsedBlocks, parsedBlock)
	}
	return parsedBlocks
}

func GetParsedBlock(filePath string) []*describe.Block {
	rawBytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	var rawData RawData
	if err = json.Unmarshal(rawBytes, &rawData); err != nil {
		panic(err)
	}
	parsedBlocks := ConvertRawData(&rawData)
	legacyBlocks := make([]*describe.Block, 0)
	for rtid, block := range parsedBlocks {
		legacyBlocks = append(legacyBlocks, describe.NewBlockFromSnbt(
			block.NameWithoutMC,
			block.States.SNBTString(),
			block.LegacyData,
			uint32(rtid),
		))
	}
	return legacyBlocks
}

func GenSpecificLegacyBlockToNemcTranslateRecords(
	c *convertor.ToNEMCConvertor,
	legacyFilePath string,
	currentBlocks *block_set.BlockSet,
) []*convertor.ConvertRecord {
	records := make([]*convertor.ConvertRecord, 0)
	legacyBlocks := GetParsedBlock(legacyFilePath)
	for _, legacyBlock := range legacyBlocks {
		legacyName := describe.BlockNameForSearch(legacyBlock.ShortName())
		legacyValue := legacyBlock.LegacyValue()
		rtid, _, found := c.TryBestSearchByState(legacyName, legacyBlock.StatesForSearch())
		if !found {
			fmt.Printf("%v %v not found\n", legacyName, legacyValue)
			continue
		}
		if currentBlocks.BlockByRtid(rtid).ShortName() == legacyBlock.ShortName() {
			continue
		}
		if legacyName.BaseName() == "wool" {
			fmt.Printf("???")
		}
		record := &convertor.ConvertRecord{
			Name:             legacyName.BaseName(),
			SNBTStateOrValue: fmt.Sprintf("%v", legacyValue),
			RTID:             rtid,
		}
		records = append(records, record)

	}
	return records
}
