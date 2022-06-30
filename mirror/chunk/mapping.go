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
var LegacyBlockToRuntimeID func(block *LegacyBlock) (runtimeID uint32)
var AirRID uint32
var NEMCAirRID uint32
var stateRuntimeIDs = map[StateHash]uint32{}
var Blocks []*GeneralBlock
var LegacyBlocks []*LegacyBlock
var LegacyRuntimeIDs = map[LegacyBlockHash]uint32{}

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
		return
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
	legacyBlocks := make([]*LegacyBlock, len(Blocks))
	for rid, _ := range Blocks {
		legacyBlocks[rid] = &LegacyBlock{Name: "", Val: 0}
	}
	for nemcRid, Rid := range nemcToMCRIDMapping {
		if legacyBlocks[Rid].Name != "" {
			continue
		}
		val := nemcToVal[nemcRid]
		if nemcRid == int(NEMCAirRID) {
			continue
		}
		legacyBlocks[Rid].Val = val
		legacyBlocks[Rid].Name = nemcToName[nemcRid]
	}
	for rid, _ := range Blocks {
		if legacyBlocks[rid].Name == "" {
			legacyBlocks[rid].Name = "air"
		}
	}
	legacyBlocks[AirRID].Name = "air"
	legacyBlocks[AirRID].Val = 0
	for rid, block := range legacyBlocks {
		LegacyRuntimeIDs[LegacyBlockHash{name: block.Name, data: block.Val}] = uint32(rid)
	}
	RuntimeIDToLegacyBlock = func(runtimeID uint32) (legacyBlock *LegacyBlock) {
		return legacyBlocks[runtimeID]
	}
}

//go:embed blockmapping_nemc_2_1_10_mc_1_19.gob.brotli
var mappingInData []byte

func init() {
	InitMapping(mappingInData)
}
