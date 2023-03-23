package items

import (
	"bytes"
	_ "embed"
	"encoding/gob"

	"github.com/andybalholm/brotli"
)

var ItemRuntimeIDToNameMapping func(rtid int32) string
var ItemRuntimeIDToItemDescribe func(rtid int32) *ItemDescribe
var RuntimeIDToItemNameMapping map[int32]*ItemDescribe

//go:embed itemRuntimeID2NameMapping_nemc_2_2_15.gob.brotli
var mappingInData []byte

type ItemDescribe struct {
	ItemName string `json:"name"`
	Meta     int    `json:"maxDamage"`
}

func init() {
	uncompressor := brotli.NewReader(bytes.NewBuffer(mappingInData))
	RuntimeIDToItemNameMapping = make(map[int32]*ItemDescribe)
	if err := gob.NewDecoder(uncompressor).Decode(&RuntimeIDToItemNameMapping); err != nil {
		panic(err)
	}
	if len(RuntimeIDToItemNameMapping) == 0 {
		panic("itemRuntimeIds read fail")
	}
	ItemRuntimeIDToNameMapping = func(rtid int32) string {
		if item := RuntimeIDToItemNameMapping[rtid]; item != nil {
			return item.ItemName
		} else {
			return ""
		}
	}
	ItemRuntimeIDToItemDescribe = func(rtid int32) *ItemDescribe {
		return RuntimeIDToItemNameMapping[rtid]
	}
}
