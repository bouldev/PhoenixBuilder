package step2_add_standard_mc_converts

import (
	"fmt"
	"phoenixbuilder/mirror/blocks/block_set"
	"phoenixbuilder/mirror/blocks/convertor"
)

func GenMCToNemcTranslateRecords(
	snbtInOut [][4]string,
	c *convertor.ToNEMCConvertor,
	blockSet *block_set.BlockSet,
) []*convertor.ConvertRecord {
	translated := 0
	ignored := 0
	conflicted := 0
	records := []*convertor.ConvertRecord{}
	postponeSnbtInOut := [][4]string{}
	for _, s := range snbtInOut {
		inBlockName, inBlockState, outBlockName, outBlockState := s[0], s[1], s[2], s[3]
		inBlockName, inBlockState, outBlockName, outBlockState = AlterInOutSnbtBlock(inBlockName, inBlockState, outBlockName, outBlockState)
		record, ok, notMatched := TryAddConvert(inBlockName, inBlockState, outBlockName, outBlockState, c, blockSet, false)
		if notMatched {
			_s := s
			postponeSnbtInOut = append(postponeSnbtInOut, _s)
			continue
		}
		if !ok {
			conflicted += 1
			continue
		}
		if record == nil {
			ignored += 1
			continue
		}
		translated += 1
		records = append(records, record)
	}
	for _, s := range postponeSnbtInOut {
		inBlockName, inBlockState, outBlockName, outBlockState := s[0], s[1], s[2], s[3]
		inBlockName, inBlockState, outBlockName, outBlockState = AlterInOutSnbtBlock(inBlockName, inBlockState, outBlockName, outBlockState)
		record, ok, _ := TryAddConvert(inBlockName, inBlockState, outBlockName, outBlockState, c, blockSet, true)
		if !ok {
			conflicted += 1
			continue
		}
		if record == nil {
			ignored += 1
			continue
		}
		translated += 1
		records = append(records, record)
	}
	fmt.Printf("translated: %v\n", translated)
	fmt.Printf("ignored: %v\n", ignored)
	fmt.Printf("conflicted: %v\n", conflicted)
	return records
}

// func PackConvertRecord(records []ConvertRecord) (raw string, compressed []byte) {
// 	textRecord := ""
// 	for _, record := range records {
// 		textRecord += fmt.Sprintf("%v\n%v\n%v\n", record.Name, record.SNBTState, record.RTID)
// 	}
// 	// os.WriteFile("../bedrock_java_to_translate.txt", []byte(textRecord), 0755)
// 	outFp := bytes.NewBuffer([]byte{})
// 	brotliWriter := brotli.NewWriter(outFp)
// 	brotliWriter.Write([]byte(textRecord))
// 	if err := brotliWriter.Close(); err != nil {
// 		panic(err)
// 	}
// 	return textRecord, outFp.Bytes()
// }
