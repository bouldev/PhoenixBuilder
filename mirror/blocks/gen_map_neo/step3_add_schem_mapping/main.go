package step3_add_schem_mapping

import (
	"encoding/json"
	"fmt"
	"phoenixbuilder/mirror/blocks/convertor"
	"phoenixbuilder/mirror/blocks/describe"
	"strings"
)

type JavaToBedrockMappingIn struct {
	Name       string         `json:"bedrock_identifier"`
	Properties map[string]any `json:"bedrock_states"`
}

func GenSchemConvertRecord(rawData []byte, c *convertor.ToNEMCConvertor) (records []*convertor.ConvertRecord) {
	var redundant = 0
	var overwrite = 0
	var translated = 0

	javaBlocks := map[string]JavaToBedrockMappingIn{}
	err := json.Unmarshal(rawData, &javaBlocks)
	if err != nil {
		panic(err)
	}
	records = make([]*convertor.ConvertRecord, 0)
	for blockIn, bedrockBlockDescribe := range javaBlocks {
		outBlockNameForSearch := describe.BlockNameForSearch(bedrockBlockDescribe.Name)
		// TODO CHECK IF THIS EXIST IN 1.19
		if strings.Contains(outBlockNameForSearch.BaseName(), "mangrove_roots") {
			continue
		}
		outBlockStateForSearch, err := describe.PropsForSearchFromNbt(bedrockBlockDescribe.Properties)
		if err != nil {
			panic(err)
		}

		rtid, found := c.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
		if !found {
			c.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
			panic("not found!")
		}

		// fmt.Println(outBlockNameForSearch, outBlockStateForSearch.InPreciseSNBT(), rtid)
		inSS := strings.Split(blockIn, "[")
		inBlockName, inBlockState := inSS[0], ""
		if len(inSS) > 1 {
			inBlockState = inSS[1]
		}
		hasWaterLoggedInfo := strings.Contains(inBlockState, "waterlogged")
		var inBlockStateForSearchWaterLogged *describe.PropsForSearch
		inBlockState = strings.TrimSuffix(inBlockState, "]")
		if hasWaterLoggedInfo {
			inBlockStateForSearchWaterLogged, err = describe.PropsForSearchFromStr(inBlockState)
			if err != nil {
				panic(err)
			}
			inBlockState = strings.ReplaceAll(inBlockState, ",waterlogged=true", "")
			inBlockState = strings.ReplaceAll(inBlockState, ",waterlogged=false", "")
			inBlockState = strings.ReplaceAll(inBlockState, "waterlogged=true,", "")
			inBlockState = strings.ReplaceAll(inBlockState, "waterlogged=false,", "")
			inBlockState = strings.ReplaceAll(inBlockState, "waterlogged=true", "")
			inBlockState = strings.ReplaceAll(inBlockState, "waterlogged=false", "")

		}

		inBlockNameForSearch := describe.BlockNameForSearch(inBlockName)
		inBlockStateForSearch, err := describe.PropsForSearchFromStr(inBlockState)
		if err != nil {
			panic(err)
		}
		if strings.HasPrefix(inBlockState, "block_data=") {
			panic("not implement")
		} else {
			if exist, err := c.AddAnchorByState(inBlockNameForSearch, inBlockStateForSearch, rtid, false); err != nil {
				overwrite++
				records = append(records, &convertor.ConvertRecord{
					Name:             inBlockNameForSearch.BaseName(),
					SNBTStateOrValue: inBlockStateForSearch.InPreciseSNBT(),
					RTID:             rtid,
				})
				if _, err := c.AddAnchorByState(inBlockNameForSearch, inBlockStateForSearch, rtid, true); err != nil {
					panic(err)
				}
			} else if exist {
				redundant++
			} else {
				translated++
				records = append(records, &convertor.ConvertRecord{
					Name:             inBlockNameForSearch.BaseName(),
					SNBTStateOrValue: inBlockStateForSearch.InPreciseSNBT(),
					RTID:             rtid,
				})
			}
			// fmt.Println(inBlockNameForSearch, inBlockStateForSearch.InPreciseSNBT(), hasWaterLoggedTrueInfo)
			if inBlockStateForSearchWaterLogged != nil {
				if exist, err := c.AddAnchorByState(inBlockNameForSearch, inBlockStateForSearchWaterLogged, rtid, false); err != nil {
					overwrite++
					records = append(records, &convertor.ConvertRecord{
						Name:             inBlockNameForSearch.BaseName(),
						SNBTStateOrValue: inBlockStateForSearchWaterLogged.InPreciseSNBT(),
						RTID:             rtid,
					})
					if _, err := c.AddAnchorByState(inBlockNameForSearch, inBlockStateForSearchWaterLogged, rtid, true); err != nil {
						panic(err)
					}
				} else if exist {
					redundant++
				} else {
					translated++
					records = append(records, &convertor.ConvertRecord{
						Name:             inBlockNameForSearch.BaseName(),
						SNBTStateOrValue: inBlockStateForSearchWaterLogged.InPreciseSNBT(),
						RTID:             rtid,
					})
				}
			}
		}
	}
	fmt.Printf("ok %v overwrite %v redundant %v\n", translated, overwrite, redundant)
	return records
}
