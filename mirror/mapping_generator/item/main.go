package main

import (
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/andybalholm/brotli"
)

//go:embed item_runtime_ids_2_1_10.json
var runtimeIDSData []byte

type ItemDesc struct {
	ItemName string `json:"name"`
	Meta     int    `json:"maxDamage"`
}

func main() {
	itemsList := map[string]*ItemDesc{}
	err := json.Unmarshal(runtimeIDSData, &itemsList)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(itemsList))
	runtimeIDToItemNameMapping := make(map[int32]*ItemDesc)
	for iStr, item := range itemsList {
		if i, err := strconv.Atoi(iStr); err != nil {
			panic(err)
		} else if item != nil {
			runtimeIDToItemNameMapping[int32(i)] = item
		}
	}
	fp, err := os.OpenFile("itemRuntimeID2NameMapping_nemc_2_1_10.gob.brotli", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	compressor := brotli.NewWriter(fp)
	if err := gob.NewEncoder(compressor).Encode(runtimeIDToItemNameMapping); err != nil {
		panic(err)
	}
	if err := compressor.Close(); err != nil {
		panic(err)
	}
	fp.Close()
}
