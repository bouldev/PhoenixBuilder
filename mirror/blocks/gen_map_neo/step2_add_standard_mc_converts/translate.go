package step2_add_standard_mc_converts

import (
	"fmt"
	"phoenixbuilder/mirror/blocks/block_set"
	"phoenixbuilder/mirror/blocks/convertor"
	"phoenixbuilder/mirror/blocks/describe"
	"strconv"
	"strings"
)

func TryAddConvert(inBlockName, inBlockState, outBlockName, outBlockState string,
	c *convertor.ToNEMCConvertor,
	blockSet *block_set.BlockSet,
	mustMatch bool) (record *convertor.ConvertRecord, ok bool, notMatched bool) {
	// first find target runtime id
	outBlockNameForSearch := describe.BlockNameForSearch(outBlockName)
	outBlockStateForSearch, err := describe.PropsForSearchFromStr(outBlockState)
	if err != nil {
		panic(err)
	}
	rtid := blockSet.UnknownRitd()
	found := false

	if strings.HasPrefix(outBlockState, "block_data=") {
		outBlockState = strings.TrimPrefix(outBlockState, "block_data=")
		blockVal, _ := strconv.Atoi(outBlockState)
		rtid, found = c.PreciseMatchByLegacyValue(outBlockNameForSearch, uint16(blockVal))
		if !found {
			if !mustMatch {
				return nil, false, true
			}
			rtid, found = c.TryBestSearchByLegacyValue(outBlockNameForSearch, uint16(blockVal))
			if !found {
				panic(fmt.Sprintf("not found! %v %v", outBlockNameForSearch, blockVal))
			} else {
				targetBlock := blockSet.BlockByRtid(rtid)
				fmt.Printf("fuzzy block data: %v %v -> %v\n", outBlockName, blockVal, targetBlock.String())
			}
		}
		if rtid == uint32(blockSet.AirRuntimeID()) {
			return nil, true, false
		}
	} else {
		rtid, found = c.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
		if !found {
			if !mustMatch {
				return nil, false, true
			}
			c.PreciseMatchByState(outBlockNameForSearch, outBlockStateForSearch)
			panic(fmt.Sprintf("not found! %v %v", outBlockNameForSearch, outBlockStateForSearch))
		}
		if rtid == uint32(blockSet.AirRuntimeID()) {
			return nil, true, false
		}
	}

	inBlockNameForSearch := describe.BlockNameForSearch(inBlockName)
	inBlockStateForSearch, err := describe.PropsForSearchFromStr(inBlockState)
	if err != nil {
		panic(err)
	}
	if strings.HasPrefix(inBlockState, "block_data=") {
		inBlockState = strings.TrimPrefix(inBlockState, "block_data=")
		blockVal, _ := strconv.Atoi(inBlockState)
		if existed, err := c.AddAnchorByLegacyValue(inBlockNameForSearch, uint16(blockVal), rtid, false); err != nil {
			// fmt.Printf("ignore %v %v -> %v orig:(%v)\n", inBlockNameForSearch, blockVal, rtid, outBlockStateForSearch.InPreciseSNBT())
			return nil, false, false
		} else if !existed {
			return &convertor.ConvertRecord{
				Name:             inBlockNameForSearch.BaseName(),
				SNBTStateOrValue: fmt.Sprintf("%v", blockVal),
				RTID:             rtid,
			}, true, false
		} else {
			return nil, true, false
		}
		// fmt.Printf("%v %v -> %v\n", inBlockNameForSearch, blockVal, rtid)
	} else {
		if existed, err := c.AddAnchorByState(inBlockNameForSearch, inBlockStateForSearch, rtid, false); err != nil {
			// fmt.Printf("ignore %v %v -> %v orig:(%v)\n", inBlockNameForSearch, inBlockStateForSearch.InPreciseSNBT(), rtid, outBlockStateForSearch.InPreciseSNBT())
			return nil, false, false
		} else if !existed {
			return &convertor.ConvertRecord{
				Name:             inBlockNameForSearch.BaseName(),
				SNBTStateOrValue: inBlockStateForSearch.InPreciseSNBT(),
				RTID:             rtid,
			}, true, false
		} else {
			return nil, true, false
		}
	}

}
