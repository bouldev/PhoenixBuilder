package items

import (
	"bytes"
	_ "embed"
	"encoding/gob"

	"github.com/andybalholm/brotli"
)

var ItemRuntimeIDToNameMapping func(rtid int32) string

//go:embed itemRuntimeID2NameMapping_nemc_1_17.gob.brotli
var mappingInData []byte

func init() {
	uncompressor := brotli.NewReader(bytes.NewBuffer(mappingInData))
	itemNames := make([]string, 0)
	if err := gob.NewDecoder(uncompressor).Decode(&itemNames); err != nil {
		panic(err)
	}
	if len(itemNames) == 0 {
		panic("itemRuntimeIds read fail")
	}
	ItemRuntimeIDToNameMapping = func(rtid int32) string {
		return itemNames[rtid]
	}
}
