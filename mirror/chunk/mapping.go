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
var blocks []*GeneralBlock
var legacyBlocks []*LegacyBlock
var legacyRuntimeIDs = map[LegacyBlockHash]uint32{}

const (
	// SubChunkVersion is the current version of the written sub chunks, specifying the format they are
	// written on disk and over network.
	SubChunkVersion = 9
	// CurrentBlockVersion is the current version of blocks (states) of the game. This version is composed
	// of 4 bytes indicating a version, interpreted as a big endian int. The current version represents
	// 1.16.0.14 {1, 16, 0, 14}.
	CurrentBlockVersion int32 = 17825806
)

// blockEntry represents a block as found in a disk save of a world.
type blockEntry struct {
	Name    string         `nbt:"name"`
	State   map[string]any `nbt:"states"`
	Version int32          `nbt:"version"`
	ID      int32          `nbt:"oldid,omitempty"` // PM writes this field, so we allow it anyway to avoid issues loading PM worlds.
	Meta    int16          `nbt:"val,omitempty"`
}

type GeneralBlock struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
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

func hashProperties(properties map[string]any) string {
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
		panic(fmt.Sprintf("cannot register the same state twice (%+v)", s))
	}
	rid := uint32(len(blocks))
	if s.Name == "minecraft:air" {
		AirRID = rid
	}
	stateRuntimeIDs[h] = rid
	blocks = append(blocks, s)
}

type MappingIn struct {
	RIDToMCBlock   []*GeneralBlock
	NEMCRidToMCRid []int16
	NEMCRidToVal   []uint8
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
	if len(blocks) == 0 {
		panic("blockStateData read fail")
	}

	RuntimeIDToState = func(runtimeID uint32) (name string, properties map[string]any, found bool) {
		if runtimeID >= uint32(len(blocks)) {
			return "", nil, false
		}
		name, properties = blocks[runtimeID].EncodeBlock()
		return name, properties, true
	}
	RuntimeIDToBlock = func(runtimeID uint32) (block *GeneralBlock, found bool) {
		if runtimeID >= uint32(len(blocks)) {
			return nil, false
		}
		return blocks[runtimeID], true
	}

	StateToRuntimeID = func(name string, properties map[string]any) (runtimeID uint32, found bool) {
		rid, ok := stateRuntimeIDs[StateHash{name: name, properties: hashProperties(properties)}]
		return rid, ok
	}

	nemcToMCRIDMapping := mappingIn.NEMCRidToMCRid

	NEMCAirRID = 134
	NEMCRuntimeIDToStandardRuntimeID = func(nemcRuntimeID uint32) (runtimeID uint32) {
		return uint32(nemcToMCRIDMapping[nemcRuntimeID])
	}
	if NEMCRuntimeIDToStandardRuntimeID(NEMCAirRID) != AirRID {
		panic("Air rid not matching")
	}

	nemcToVal := mappingIn.NEMCRidToVal
	legacyBlocks := make([]*LegacyBlock, len(blocks))
	for rid, block := range blocks {
		legacyBlocks[rid] = &LegacyBlock{Name: block.Name, Val: 0}
	}
	for nemcRid, Rid := range nemcToMCRIDMapping {
		val := nemcToVal[nemcRid]
		if nemcRid == int(NEMCAirRID) {
			continue
		}
		legacyBlocks[Rid].Val = val
	}
	for rid, block := range legacyBlocks {
		legacyRuntimeIDs[LegacyBlockHash{name: block.Name, data: block.Val}] = uint32(rid)
	}
	RuntimeIDToLegacyBlock = func(runtimeID uint32) (legacyBlock *LegacyBlock) {
		return legacyBlocks[runtimeID]
	}
}

//go:embed blockmapping_nemc_2_1_10_mc_1_18_30.gob.brotli
var mappingInData []byte

func init() {
	InitMapping(mappingInData)
}
