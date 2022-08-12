package structure

import (
	"fmt"
	"os"
	"path"
	"phoenixbuilder/mirror"
	"phoenixbuilder/mirror/define"
	"phoenixbuilder/mirror/io/mcdb"
	"phoenixbuilder/omega/utils"
	"time"

	"github.com/df-mc/goleveldb/leveldb/opt"
)

func EncodeMCWorld(chunks map[define.ChunkPos]*mirror.ChunkData, startPos, endPos define.CubePos, structureName string, targetDir string) (err error) {
	provider, err := mcdb.New(path.Join(targetDir, structureName), opt.FlateCompression)
	if err != nil {
		return err
	}
	provider.D.LevelName = structureName
	for _, cd := range chunks {
		provider.Write(cd)
	}
	fp, err := os.OpenFile(path.Join(targetDir, structureName, "Omega导出建筑记录.txt"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	fmt.Fprintf(fp, "@STRUCTURE: %v @START: %v %v %v @END: %v %v %v @TIME: %v\n", structureName, startPos.X(), startPos.Y(), startPos.Z(), endPos.X(), endPos.Y(), endPos.Z(), utils.TimeToString(time.Now()))
	fp.Close()
	provider.Close()

	structureNameWithPos := fmt.Sprintf("%v@%v,%v,%v.mcworld", structureName, startPos.X(), startPos.Y(), startPos.Z())
	fp, err = os.OpenFile(structureNameWithPos, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	err = utils.Zip(path.Join(targetDir, structureName), fp, func(filePath string, info os.FileInfo) (discard bool) { return false })
	if err != nil {
		return err
	}
	fp.Close()
	return
}
