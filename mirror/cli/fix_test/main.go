package main

import (
	_ "embed"
	"fmt"
	"os"
	"phoenixbuilder/mirror/chunk"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/mcdb"
	"sort"

	"github.com/df-mc/goleveldb/leveldb/opt"
)

type PosCombinedResult struct {
	pos  define.ChunkPos
	info *mcdb.ChunksInfo
}
type PosCombinedResults []*PosCombinedResult

func (p *PosCombinedResults) Len() int {
	return len(*p)
}

func (p *PosCombinedResults) Less(i, j int) bool {
	if (*p)[i].pos[0] != (*p)[j].pos[0] {
		return (*p)[i].pos[0] < (*p)[j].pos[0]
	} else {
		return (*p)[i].pos[1] < (*p)[j].pos[1]
	}
}

func (p *PosCombinedResults) Swap(i, j int) {
	k := (*p)[i]
	(*p)[i] = (*p)[j]
	(*p)[j] = k
}

func main() {
	in_provider, err := mcdb.New("C:/Users/dai/AppData/Local/Packages/Microsoft.MinecraftUWP_8wekyb3d8bbwe/LocalState/games/com.mojang/minecraftWorlds/1037349583", opt.FlateCompression)
	if err != nil {
		panic(err)
	}
	result := in_provider.IterAll()
	// for pos, r := range result {
	// 	fmt.Println(pos)
	// 	fmt.Println(*r)
	// }
	resultsList := PosCombinedResults{}
	for pos, r := range result {
		resultsList = append(resultsList, &PosCombinedResult{pos, r})
	}
	sort.Sort(&resultsList)
	// fmt.Println(resultsList)
	numChunks := len(resultsList)
	fmt.Println(numChunks)

	splitChunks := PosCombinedResults{}
	for _, v := range resultsList {
		// if i != 223 {
		// 	continue
		// }
		splitChunks = append(splitChunks, v)
	}

	errChunkPos := splitChunks[0].pos
	fmt.Println(errChunkPos)
	chunkData := in_provider.Get(errChunkPos)
	for i, subChunk := range chunkData.Chunk.Sub() {
		fmt.Println("sub_chunk: ", i, subChunk.Empty())
		Palette := subChunk.Layer(0).Palette()
		if !subChunk.Empty() {
			for i := uint16(0); int(i) < Palette.Len(); i++ {
				rid := Palette.Value(i)
				fmt.Println(chunk.RuntimeIDToState(rid))
			}
			for x := byte(0); x < 16; x++ {
				for y := byte(0); y < 16; y++ {
					for z := byte(0); z < 16; z++ {
						if subChunk.Block(x, y, z, 0) == chunk.AirRID {
							continue
						} else {
							// if i == 5 {
							// 	subChunk.SetBlock(x, y, z, 0, chunk.AirRID)
							// }
							// blockName, prop, _ := chunk.RuntimeIDToState(subChunk.Block(x, y, z, 0))
							// fmt.Println(x, y, z, blockName, prop)
						}
					}
				}
			}
		}
	}
	// fmt.Println(chunkData.Chunk.Sub())

	reducedChunks := make(map[define.ChunkPos]*mcdb.ChunksInfo)
	for _, r := range splitChunks {
		reducedChunks[r.pos] = r.info
	}
	fmt.Println(reducedChunks)

	out_path := "C:/Users/dai/AppData/Local/Packages/Microsoft.MinecraftUWP_8wekyb3d8bbwe/LocalState/games/com.mojang/minecraftWorlds/fix_out"
	os.RemoveAll(out_path)
	out_provider, err := mcdb.New(out_path, opt.FlateCompression)
	if err != nil {
		panic(err)
	}
	{
		iter := in_provider.DB.NewIterator(nil, nil)
		for iter.Next() {
			key := iter.Key()
			if len(key) < 8 {
				continue
			}
			pos, _ := in_provider.Position(key)
			if _, hasK := reducedChunks[pos]; hasK {
				// if pos == errChunkPos {
				data, err := in_provider.DB.Get(key, nil)
				// in_provider.Get()
				if err != nil {
					panic(err)
				}
				out_provider.DB.Put(key, data, nil)
			}

		}
	}
	// out_provider.D = in_provider.D
	out_provider.D.LevelName = "fix_out"
	out_provider.Close()
}
