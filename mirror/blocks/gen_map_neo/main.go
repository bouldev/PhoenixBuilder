package main

import (
	"bytes"
	"os"
	"phoenixbuilder/mirror/blocks/block_set"
	"phoenixbuilder/mirror/blocks/convertor"
	"phoenixbuilder/mirror/blocks/gen_map_neo/step0_nemc_blocks_liliya"
	"phoenixbuilder/mirror/blocks/gen_map_neo/step2_add_specific_legacy_converts"
	"phoenixbuilder/mirror/blocks/gen_map_neo/step2_add_standard_mc_converts"
	"phoenixbuilder/mirror/blocks/gen_map_neo/step3_add_schem_mapping"

	"github.com/andybalholm/brotli"
)

func main() {
	readBlocks := step0_nemc_blocks_liliya.GetParsedBlock("./data/block_palette_2.12.json")
	blocksRecordString := readBlocks.DumpStringRecords()
	// CHECK
	{
		// fmt.Println(blocksRecordString)
		reLoadedBlocks := block_set.BlockSetFromStringRecords(blocksRecordString, 0xffffffff)
		reDumpedBlocksRecordString := reLoadedBlocks.DumpStringRecords()
		if reDumpedBlocksRecordString != blocksRecordString {
			panic("mismatch")
		}
		// b0 := reLoadedBlocks.BlockByRtid(0)
		// fmt.Println(b0)
		// b3736 := reLoadedBlocks.BlockByRtid(3736)
		// // cobblestone_wall {wall_block_type:"mossy_stone_brick",wall_connection_type_east:"tall",wall_connection_type_north:"none",wall_connection_type_south:"tall",wall_connection_type_west:"short",wall_post_bit:1b} (Value: 3352) (RuntimeID: 3736)
		// fmt.Println(b3736)
		// b13142 := reLoadedBlocks.BlockByRtid(13142)
		// fmt.Println(b13142)
	}
	// SAVE
	{
		outBuf := bytes.NewBuffer([]byte{})
		brotliWriter := brotli.NewWriter(outBuf)
		brotliWriter.Write([]byte(blocksRecordString))
		if err := brotliWriter.Close(); err != nil {
			panic(err)
		}
		if err := os.WriteFile("nemc.br", outBuf.Bytes(), 0755); err != nil {
			panic(err)
		}
	}
	// Generate Standard MC converts
	standardMCConvertsRecordsString := ""
	{
		convertorForMC := readBlocks.CreateEmptyConvertor()
		snbtInOut := step2_add_standard_mc_converts.ReadSnbtFile("./data/snbt_convert.txt")
		convertRecords := step2_add_standard_mc_converts.GenMCToNemcTranslateRecords(
			snbtInOut,
			convertorForMC,
			&readBlocks,
		)

		for _, r := range convertRecords {
			standardMCConvertsRecordsString += r.String()
		}
		//CHECK
		{
			reLoadRecords, err := convertor.ReadRecordsFromString(standardMCConvertsRecordsString)
			if err != nil {
				panic(err)
			}
			reDumpRecords := ""
			for _, r := range reLoadRecords {
				reDumpRecords += r.String()
			}
			if reDumpRecords != standardMCConvertsRecordsString {
				panic("mismatch")
			}
		}
		// SAVE
		{
			outBuf := bytes.NewBuffer([]byte{})
			brotliWriter := brotli.NewWriter(outBuf)
			brotliWriter.Write([]byte(standardMCConvertsRecordsString))
			if err := brotliWriter.Close(); err != nil {
				panic(err)
			}
			if err := os.WriteFile("bedrock_java_to_translate.br", outBuf.Bytes(), 0755); err != nil {
				panic(err)
			}
		}
	}
	// Generate Specific Version Legacy Value Converts
	specificVersionLegacyValueConvertsRecordString := ""
	{
		convertForSpecificVersion := readBlocks.CreateEmptyConvertor()
		// Load Standard MC Records
		{
			standardMCConvertsRecord, err := convertor.ReadRecordsFromString(standardMCConvertsRecordsString)
			if err != nil {
				panic(err)
			}
			for _, r := range standardMCConvertsRecord {
				convertForSpecificVersion.LoadConvertRecord(r, false, true)
			}
		}
		convertRecords := step2_add_specific_legacy_converts.GenSpecificLegacyBlockToNemcTranslateRecords(convertForSpecificVersion, "data/legacy_block_palette_2.11.json", &readBlocks)
		for _, r := range convertRecords {
			specificVersionLegacyValueConvertsRecordString += r.String()
		}
		// SAVE
		{
			outBuf := bytes.NewBuffer([]byte{})
			brotliWriter := brotli.NewWriter(outBuf)
			brotliWriter.Write([]byte(specificVersionLegacyValueConvertsRecordString))
			if err := brotliWriter.Close(); err != nil {
				panic(err)
			}
			if err := os.WriteFile("specific_legacy_value_to_translate.br", outBuf.Bytes(), 0755); err != nil {
				panic(err)
			}
		}
	}
	// Generate Schem Converts
	{
		convertForSchem := readBlocks.CreateEmptyConvertor()
		// Load Standard MC Records
		{
			standardMCConvertsRecord, err := convertor.ReadRecordsFromString(standardMCConvertsRecordsString)
			if err != nil {
				panic(err)
			}
			for _, r := range standardMCConvertsRecord {
				convertForSchem.LoadConvertRecord(r, false, true)
			}
		}
		rawSchemData, err := os.ReadFile("./data/block_1_18_java_to_bedrock.json")
		if err != nil {
			panic(err)
		}
		convertRecords := step3_add_schem_mapping.GenSchemConvertRecord(
			rawSchemData,
			convertForSchem,
		)
		ConvertsRecordsString := ""
		for _, r := range convertRecords {
			ConvertsRecordsString += r.String()
		}
		//CHECK
		{
			reLoadRecords, err := convertor.ReadRecordsFromString(ConvertsRecordsString)
			if err != nil {
				panic(err)
			}
			reDumpRecords := ""
			for _, r := range reLoadRecords {
				reDumpRecords += r.String()
			}
			if reDumpRecords != ConvertsRecordsString {
				panic("mismatch")
			}
		}
		// SAVE
		{
			outBuf := bytes.NewBuffer([]byte{})
			brotliWriter := brotli.NewWriter(outBuf)
			brotliWriter.Write([]byte(ConvertsRecordsString))
			if err := brotliWriter.Close(); err != nil {
				panic(err)
			}
			if err := os.WriteFile("schem_to_translate.br", outBuf.Bytes(), 0755); err != nil {
				panic(err)
			}
		}
	}
}
