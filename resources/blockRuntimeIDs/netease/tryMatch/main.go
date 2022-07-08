package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"phoenixbuilder/minecraft/nbt"
)

//go:embed block_states.nbt
var blockStateData []byte

type GeneralBlock struct {
	Name       string         `nbt:"name"`
	Properties map[string]any `nbt:"states"`
	Version    int32          `nbt:"version"`
}

func main() {
	dec := nbt.NewDecoder(bytes.NewBuffer(blockStateData))
	out_val := make([][2]interface{}, 0)
	lastName := ""
	dataIdx := 0
	rid := 0
	for {

		var s GeneralBlock
		if err := dec.Decode(&s); err != nil {
			break
		}
		s.Name = s.Name[10:]
		if s.Name == "frog_egg" || s.Name == "verdant_froglight" {
			fmt.Println(rid)
			continue
		}
		if s.Name == "skull" {
			if lastName == "skull" {
				continue
			} else {
				lastName = "skull"
				for i := 0; i < 12; i++ {
					out_val = append(out_val, [2]interface{}{
						fmt.Sprintf("skull"),
						dataIdx,
						// rid,
					})
					rid++
				}
				continue
			}
		}

		if s.Name != lastName {
			dataIdx = 0
			lastName = s.Name
		} else {
			dataIdx++
		}
		out_val = append(out_val, [2]interface{}{
			fmt.Sprintf("%v", s.Name),
			dataIdx,
			// rid,
		})
		rid++
	}
	fp, err := os.OpenFile("compare.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(fp)
	enc.SetIndent("", "\t")
	enc.Encode(out_val)
}
