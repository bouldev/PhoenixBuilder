package blocks

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"phoenixbuilder/mirror/blocks/block_set"
	"phoenixbuilder/mirror/blocks/convertor"
	"phoenixbuilder/mirror/blocks/describe"

	"github.com/andybalholm/brotli"
)

//go:embed "nemc.br"
var nemcBlockInfoBytes []byte

//go:embed "bedrock_java_to_translate.br"
var toNemcDataLoadBedrockJavaTranslateInfo []byte

//go:embed "specific_legacy_value_to_translate.br"
var toNemcDataLoadSpecificLegacyValuesTranslateInfo []byte

//go:embed "schem_to_translate.br"
var toNemcDataLoadSchemTranslateInfo []byte

func readAndUnpack(bs []byte) string {
	dataBytes, err := io.ReadAll(brotli.NewReader(bytes.NewBuffer(bs)))
	if err != nil {
		panic(err)
	}
	return string(dataBytes)
}

var MC_CURRENT *block_set.BlockSet
var MC_1_20_10 *block_set.BlockSet

// duplicate in future
var NEMC_BLOCK_VERSION = uint32(0)
var NEMC_AIR_RUNTIMEID = uint32(0)
var AIR_RUNTIMEID = uint32(0)

func initNEMCBlocks() {
	bs := block_set.BlockSetFromStringRecords(readAndUnpack(nemcBlockInfoBytes), 0xFFFFFFFF)
	MC_CURRENT = bs
	MC_1_20_10 = bs
	NEMC_BLOCK_VERSION = bs.Version()
	NEMC_AIR_RUNTIMEID = bs.AirRuntimeID()
	AIR_RUNTIMEID = bs.AirRuntimeID()
}

var DefaultAnyToNemcConvertor *convertor.ToNEMCConvertor
var SchemToNemcConvertor *convertor.ToNEMCConvertor
var quickSchematicMapping [256][16]uint32

func initSchematicBlockCheck(schematicToNemcConvertor *convertor.ToNEMCConvertor) {
	quickSchematicMapping = [256][16]uint32{}
	for i := 0; i < 256; i++ {
		blockName := schematicBlockStrings[i]
		_, found := DefaultAnyToNemcConvertor.TryBestSearchByLegacyValue(describe.BlockNameForSearch(blockName), 0)
		if !found {
			panic(fmt.Errorf("schematic %v 0 not found", blockName))
		}
	}
	for blockI := 0; blockI < 256; blockI++ {
		blockName := schematicBlockStrings[blockI]
		// if blockName == "stone_slab" {
		// 	fmt.Println("slab")
		// }
		for dataI := 0; dataI < 16; dataI++ {
			rtid, found := schematicToNemcConvertor.TryBestSearchByLegacyValue(describe.BlockNameForSearch(blockName), uint16(dataI))
			if !found || rtid == AIR_RUNTIMEID {
				rtid, _ = schematicToNemcConvertor.TryBestSearchByLegacyValue(describe.BlockNameForSearch(blockName), 0)
			}
			quickSchematicMapping[blockI][dataI] = rtid
		}
	}
	schematicToNemcConvertor = nil
}

func initConvertor() {
	DefaultAnyToNemcConvertor = MC_CURRENT.CreateEmptyConvertor()
	SchemToNemcConvertor = MC_CURRENT.CreateEmptyConvertor()
	schematicToNemcConvertor := MC_CURRENT.CreateEmptyConvertor()
	mcConvertRecords, err := convertor.ReadRecordsFromString(readAndUnpack(toNemcDataLoadBedrockJavaTranslateInfo))
	if err != nil {
		panic(err)
	}
	specificLegacyValuesConvertRecords, err := convertor.ReadRecordsFromString(readAndUnpack(toNemcDataLoadSpecificLegacyValuesTranslateInfo))
	if err != nil {
		panic(err)
	}
	for _, r := range mcConvertRecords {
		DefaultAnyToNemcConvertor.LoadConvertRecord(r, false, true)
		SchemToNemcConvertor.LoadConvertRecord(r, false, true)
		schematicToNemcConvertor.LoadConvertRecord(r, false, true)
	}
	for _, r := range specificLegacyValuesConvertRecords {
		DefaultAnyToNemcConvertor.LoadConvertRecord(r, true, true)
		SchemToNemcConvertor.LoadConvertRecord(r, true, true)
		schematicToNemcConvertor.LoadConvertRecord(r, true, true)
	}
	schemConvertRecords, err := convertor.ReadRecordsFromString(readAndUnpack(toNemcDataLoadSchemTranslateInfo))
	if err != nil {
		panic(err)
	}
	for _, r := range schemConvertRecords {
		DefaultAnyToNemcConvertor.LoadConvertRecord(r, false, false)
		SchemToNemcConvertor.LoadConvertRecord(r, true, true)
	}
	initSchematicBlockCheck(schematicToNemcConvertor)
}

func init() {
	initNEMCBlocks()
	initConvertor()
}

// var DefaultAnyToNemcConvertor = NewToNEMCConverter()
// var SchemToNemcConvertor = NewToNEMCConverter()
// var schematicToNemcConvertor = NewToNEMCConverter()

// const UNKNOWN_RUNTIME = uint32(0xFFFFFFFF)

// var NEMC_BLOCK_VERSION = int32(0)
// var NEMC_AIR_RUNTIMEID = uint32(0)
// var AIR_RUNTIMEID = uint32(0)
// var AIR_BLOCK = &NEMCBlock{
// 	Name:  "air",
// 	Value: 0,
// 	Props: make(Props, 0),
// }

// var nemcBlocks = []NEMCBlock{}

// func initNemcBlocks() {
// 	dataBytes, err := io.ReadAll(brotli.NewReader(bytes.NewBuffer(nemcBlockInfoBytes)))
// 	if err != nil {
// 		panic(err)
// 	}
// 	LoadNemcBlocksToGlobal(string(dataBytes))
// }

// var quickSchematicMapping [256][16]uint32

// func initSchematicBlockCheck() {
// 	quickSchematicMapping = [256][16]uint32{}
// 	for i := 0; i < 256; i++ {
// 		blockName := schematicBlockStrings[i]
// 		_, found := DefaultAnyToNemcConvertor.TryBestSearchByLegacyValue(BlockNameForSearch(blockName), 0)
// 		if !found {
// 			panic(fmt.Errorf("schematic %v 0 not found", blockName))
// 		}
// 	}
// 	for blockI := 0; blockI < 256; blockI++ {
// 		blockName := schematicBlockStrings[blockI]
// 		// if blockName == "stone_slab" {
// 		// 	fmt.Println("slab")
// 		// }
// 		for dataI := 0; dataI < 16; dataI++ {
// 			rtid, found := schematicToNemcConvertor.TryBestSearchByLegacyValue(BlockNameForSearch(blockName), int16(dataI))
// 			if !found || rtid == AIR_RUNTIMEID {
// 				rtid, _ = schematicToNemcConvertor.TryBestSearchByLegacyValue(BlockNameForSearch(blockName), 0)
// 			}
// 			quickSchematicMapping[blockI][dataI] = rtid
// 		}
// 	}
// 	schematicToNemcConvertor = nil
// }

// func initToNemcDataLoadBedrockJava() {
// 	WriteNemcInfoToConvertor(DefaultAnyToNemcConvertor)
// 	WriteNemcInfoToConvertor(SchemToNemcConvertor)
// 	WriteNemcInfoToConvertor(schematicToNemcConvertor)
// 	dataBytes, err := io.ReadAll(brotli.NewReader(bytes.NewBuffer(toNemcDataLoadBedrockJavaTranslateInfo)))
// 	if err != nil {
// 		panic(err)
// 	}
// 	records := string(dataBytes)
// 	DefaultAnyToNemcConvertor.LoadConvertRecords(records, false, true)
// 	SchemToNemcConvertor.LoadConvertRecords(records, false, true)
// 	schematicToNemcConvertor.LoadConvertRecords(records, false, true)
// }

// func initToNemcDataLoadSchem() {
// 	dataBytes, err := io.ReadAll(brotli.NewReader(bytes.NewBuffer(toNemcDataLoadSchemTranslateInfo)))
// 	if err != nil {
// 		panic(err)
// 	}
// 	records := string(dataBytes)
// 	DefaultAnyToNemcConvertor.LoadConvertRecords(records, false, false)
// 	SchemToNemcConvertor.LoadConvertRecords(records, true, true)
// }

// func init() {
// 	initNemcBlocks()
// 	initToNemcDataLoadBedrockJava()
// 	initToNemcDataLoadSchem()
// 	initSchematicBlockCheck()
// }
