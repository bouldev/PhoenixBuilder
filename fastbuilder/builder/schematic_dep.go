// +build dep___do_not_add_this_tag_

package builder

import (
	"compress/gzip"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/pterm/pterm"
	"io/ioutil"
	bridge_path "phoenixbuilder/fastbuilder/builder/path"
	"phoenixbuilder/fastbuilder/i18n"
	"phoenixbuilder/fastbuilder/types"
)

/*
#cgo !windows,!darwin,!ios,!ish,!android LDFLAGS: -L${SRCDIR}/../../depends/stub -lz
#cgo windows CFLAGS: -I${SRCDIR}/../../depends/zlib-1.2.12
#cgo LDFLAGS: -L${SRCDIR}/../../depends/zlib-1.2.12/prebuilt
#cgo windows,amd64 LDFLAGS: -lz-x86_64-windows -lws2_32
#cgo windows,386 LDFLAGS: -lz-i686-windows -lws2_32
#cgo darwin,!ios,amd64 LDFLAGS: -lz-x86_64-macos
#cgo darwin,!ios,arm64 LDFLAGS: -lz-arm64-macos
#cgo ios,arm64 LDFLAGS: -lz-arm64-ios
#cgo ish LDFLAGS: -lz-ish
#cgo android,arm LDFLAGS: -lz-armv7a-android
#cgo android,arm64 LDFLAGS: -lz-arm64-android
#cgo android,386 LDFLAGS: -lz-i686-android
#cgo android,amd64 LDFLAGS: -lz-x86_64-android
#cgo !windows CFLAGS: -I${SRCDIR}/../../depends/zlib-1.2.12
#include <stdint.h>
extern const char *zlibVersion(void);
extern unsigned char builder_schematic_process_schematic_file(uint32_t channelID, char *path, int64_t beginX, int64_t beginY, int64_t beginZ);
*/
import "C"

var lastChannelID uint=0
var channelMap map[uint]chan *types.Module=map[uint]chan *types.Module{}

//export builder_schematic_channel_input
func builder_schematic_channel_input(channelID uint32, x int64, y int64, z int64, id uint8, data uint8) {
	var b types.Block
	b.Name = &BlockStr[int(id)]
	b.Data = uint16(data)
	blc:=channelMap[uint(channelID)]
	blc <- &types.Module{Point: types.Position{int(x),int(y),int(z)}, Block: &b}
}

func Schematic(config *types.MainConfig, blc chan *types.Module) error {
	// Check zlib version
	zlib_safe, err := version.NewVersion("1.2.12")
	zlib_current, err := version.NewVersion(C.GoString(C.zlibVersion()))
	if zlib_current.LessThan(zlib_safe) {
		pterm.Println(pterm.Yellow(I18n.T(I18n.Notice_ZLIB_CVE)))
	}

	channelMap[lastChannelID]=blc
	gotChannelID:=lastChannelID
	lastChannelID++
	retval:=C.builder_schematic_process_schematic_file(C.uint32_t(gotChannelID), C.CString(config.Path), C.int64_t(config.Position.X), C.int64_t(config.Position.Y), C.int64_t(config.Position.Z))
	delete(channelMap, gotChannelID)
	fmt.Printf("RET %d\n",retval)
	return nil
	file, err:=bridge_path.ReadFile(config.Path)
	if err != nil {
		return I18n.ProcessSystemFileError(err)
	}
	defer file.Close()
	gzip, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzip.Close()
	
	buffer, err := ioutil.ReadAll(gzip)

	var SchematicModule struct {
		Blocks    []byte `nbt:"Blocks"`
		Data      []byte `nbt:"Data"`
		Width     int    `nbt:"Width"`
		Length    int    `nbt:"Length"`
		Height    int    `nbt:"Height"`
		WEOffsetX int    `nbt:"WEOffsetX"`
		WEOffsetY int    `nbt:"WEOffsetY"`
		WEOffsetZ int    `nbt:"WEOffsetZ"`
	}

	if err := nbt.Unmarshal(buffer, &SchematicModule); err != nil {
		// Won't return the error `err` since it contains a large content that can 
		// crash the server after being sent.
		return fmt.Errorf(I18n.T(I18n.Sch_FailedToResolve))
	}
	if(len(SchematicModule.Blocks)==0) {
		return fmt.Errorf("Invalid structure.")
	}
	Size := [3]int{SchematicModule.Width, SchematicModule.Height, SchematicModule.Length}
	Offset := [3]int{SchematicModule.WEOffsetX, SchematicModule.WEOffsetY, SchematicModule.WEOffsetZ}
	X, Y, Z := 0, 1, 2
	BlockIndex := 0

	for y := 0; y < Size[Y]; y++ {
		for z := 0; z < Size[Z]; z++ {
			for x := 0; x < Size[X]; x++ {
				p := config.Position
				p.X += x + Offset[X]
				p.Y += y + Offset[Y]
				p.Z += z + Offset[Z]
				var b types.Block
				b.Name = &BlockStr[SchematicModule.Blocks[BlockIndex]]
				b.Data = uint16(SchematicModule.Data[BlockIndex])
				if BlockIndex - 188 <= 5 && BlockIndex - 188 >= 0 {
					b.Name = &FenceName
					b.Data = uint16(BlockIndex - 188)
				}
				if BlockIndex == 3 && b.Data == 2 {
					b.Name = &PodzolName
				}
				if *b.Name != "air" {
					blc <- &types.Module{Point: p, Block: &b}
				}
				BlockIndex++
			}
		}
	}
	return nil
}
