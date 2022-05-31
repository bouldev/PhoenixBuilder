package main

import (
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"

	"github.com/andybalholm/brotli"
)

//go:embed runtimeids.json
var runtimeIDSData []byte

type ItemDescribe struct {
	ItemName string `json:"name"`
	Meta     int    `json:"maxDamage"`
}

func main() {
	itemsList := []*ItemDescribe{}
	err := json.Unmarshal(runtimeIDSData, &itemsList)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(itemsList))
	runtimeIDToItemNameMapping := make([]string, len(itemsList))
	for i, item := range itemsList {
		fmt.Println(i, item)
		if item != nil {
			runtimeIDToItemNameMapping[i] = item.ItemName
		} else {
			runtimeIDToItemNameMapping[i] = "undefined"
		}

	}
	fp, err := os.OpenFile("itemRuntimeID2NameMapping_nemc_1_17.gob.brotli", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
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
