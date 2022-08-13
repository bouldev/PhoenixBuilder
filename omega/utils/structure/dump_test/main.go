package main

import (
	"fmt"
	"os"
	"path"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/mcdb"
	"phoenixbuilder/omega/utils/structure"
	"strconv"
	"strings"

	_ "embed"

	"github.com/df-mc/goleveldb/leveldb/opt"
)

func ParseSaveCmd(cmds []string) (startPos, endPos define.CubePos, structureName string, err error) {
	paramsNames := []string{"[建筑名]", "[起点x]", "[起点y]", "[起点z]", "[终点x]", "[终点y]", "[终点z]"}
	err = fmt.Errorf("保存指令错误, 应该为:\n %v "+strings.Join(paramsNames, " "), cmds[0])
	values := [6]int{}
	if len(cmds) < 8 {
		return
	}
	structureName = cmds[1]
	for i := 0; i < 6; i++ {
		if v, _err := strconv.Atoi(cmds[i+2]); _err != nil {
			err = fmt.Errorf(err.Error() + fmt.Sprintf("\n参数 %v 不正确", paramsNames[i+1]))
		} else {
			values[i] = v
		}
	}
	sortStart := func(i int) {
		startPos[i] = values[i]
		endPos[i] = values[i+3]
		if values[i] > values[i+3] {
			startPos[i] = values[i+3]
			endPos[i] = values[i]
		}
	}
	for i := 0; i < 3; i++ {
		sortStart(i)
	}
	err = nil
	return
}

func AcquireData(hopPath *structure.ExportHopPosMap, allRequiredChunks *structure.ExportedChunksMap) (map[define.ChunkPos]*mirror.ChunkData, error) {
	provider, err := mcdb.New("/Users/dai/projects/PhoenixBuilder/omega_storage/worlds/current", opt.FlateCompression)
	if err != nil {
		return nil, err
	}
	chunks := make(map[define.ChunkPos]*mirror.ChunkData)
	for pos, _ := range *allRequiredChunks {
		if cd := provider.Get(pos); cd != nil {
			// fmt.Println(cd.BlockNbts)
			chunks[pos] = cd
			allRequiredChunks.Hit(pos)
		} else {
			// return nil, fmt.Errorf("chunk %v (Pos: %v) not found", pos, define.ChunkPos{pos.X() << 4, pos.Z() << 4})
		}
	}
	if len(*allRequiredChunks) > 0 {
		return nil, fmt.Errorf("cannot obtain all chunks")
	}
	return chunks, nil
}

type SchemFileStructrue struct {
	Palette       map[string]int32
	Metadata      map[string]interface{}
	DataVersion   int32
	BlockDataIn   []byte      `nbt:"BlockData"`
	OffsetIn      interface{} `nbt:"Offset"`
	blockData     []uint16
	offset        []int
	PaletteMax    int32
	Version       int32
	Length        int16
	Height        int16
	Width         int16
	BlockEntities []map[string]interface{}
}

//go:embed 3.schem
var data []byte

func main() {

	_, _, _, _, err := structure.DecodeSchem(data, func(s string) {})
	if err != nil {
		panic(err)
	}

	cmdStr := "save test -3 -5 7 106 205 137"
	cmds := strings.Split(cmdStr, " ")
	startPos, endPos, structureName, err := ParseSaveCmd(cmds)
	if err != nil {
		panic(err)
	}
	hopPath, allRequiredChunks := structure.PlanHopSwapPath(startPos.X(), startPos.Z(), endPos.X(), endPos.Z(), 4)

	fmt.Println(hopPath)
	fmt.Println(allRequiredChunks)
	fmt.Println(structureName)

	chunks, err := AcquireData(hopPath, allRequiredChunks)
	if err != nil {
		panic(err)
	}

	overallCacheDir := path.Join("omega_export_cache", structureName)
	if err := os.MkdirAll(overallCacheDir, 0755); err != nil {
		panic(err)
	}

	err = structure.EncodeMCWorld(chunks, startPos, endPos, structureName, overallCacheDir)
	if err != nil {
		panic(err)
	}
	err = structure.EncodeSchem(chunks, startPos, endPos, structureName, overallCacheDir)
	if err != nil {
		panic(err)
	}
}
