package items

import (
	"bytes"
	_ "embed"
	"encoding/gob"

	"github.com/andybalholm/brotli"
)

var ItemRuntimeIDToNameMapping func(rtid int32) string
var ItemRuntimeIDToItemDescribe func(rtid int32) *ItemDescribe

//go:embed itemRuntimeID2NameMapping_nemc_2_1_10.gob.brotli
var mappingInData []byte

type ItemDescribe struct {
	ItemName string `json:"name"`
	Meta     int    `json:"maxDamage"`
}

func init() {
	uncompressor := brotli.NewReader(bytes.NewBuffer(mappingInData))
	runtimeIDToItemNameMapping := make(map[int32]*ItemDescribe)
	if err := gob.NewDecoder(uncompressor).Decode(&runtimeIDToItemNameMapping); err != nil {
		panic(err)
	}
	if len(runtimeIDToItemNameMapping) == 0 {
		panic("itemRuntimeIds read fail")
	}
	ItemRuntimeIDToNameMapping = func(rtid int32) string {
		if item := runtimeIDToItemNameMapping[rtid]; item != nil {
			return item.ItemName
		} else {
			return ""
		}
	}
	ItemRuntimeIDToItemDescribe = func(rtid int32) *ItemDescribe {
		return runtimeIDToItemNameMapping[rtid]
	}
}
