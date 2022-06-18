// +build is_tweak

package special_tasks

import (
	"fmt"
	"phoenixbuilder/fastbuilder/bdump"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/environment"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/types"
	"runtime"
	"strings"
	"unsafe"
)

/*
#include <stdlib.h>

struct ranged_result {
	// e.g. grass\0tnt\0dirt\0\0
	char *palette;
	unsigned char *blocks_map;
	unsigned short *data;
	// e.g. \x03\x1e\x3e\x00[NBT Data, length=0x3e]\x02
	char *nbt_area;
};

struct single_nbt_data {
	unsigned short length;
	char *ptr;
};

// CALLER free. 
// return (struct ranged_result *)3 if not implemented but phoenixbuilder_get_block is implemented.
// return NULL if all of them are not implemented.
struct ranged_result *phoenixbuilder_get_ranged_blocks(int begin_x, int begin_y, int begin_z,
						int end_x, int end_y, int end_z);
// If phoenixbuilder_get_ranged_blocks is available,
// methods below will never be called.
char *phoenixbuilder_get_block(int x, int y, int z);
unsigned short phoenixbuilder_get_block_data(int x, int y, int z);
struct single_nbt_data *phoenixbuilder_get_block_nbt(int x, int y, int z);
*/
import "C"


type SolidSimplePos struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
	Z int64 `json:"z"`
}

type SolidRet struct {
	BlockName string `json:"blockName"`
	Position SolidSimplePos `json:"position"`
	StatusCode int64 `json:"statusCode"`
}

var ExportWaiter chan map[string]interface{}

func CreateExportTask(commandLine string, env *environment.PBEnvironment) *task.Task {
	cmdsender:=env.CommandSender
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig(env).Main())
	if err!=nil {
		cmdsender.Output(fmt.Sprintf("Failed to parse command: %v",err))
		return nil
	}
	beginPos := cfg.Position
	endPos := cfg.End
	startX,endX,startZ,endZ:=0,0,0,0
	if(endPos.X-beginPos.X<0) {
		temp:=endPos.X
		endPos.X=beginPos.X
		beginPos.X=temp
	}
	startX,endX=beginPos.X,endPos.X
	if(endPos.Y-beginPos.Y<0) {
		temp:=endPos.Y
		endPos.Y=beginPos.Y
		beginPos.Y=temp
	}
	if(endPos.Z-beginPos.Z<0) {
		temp:=endPos.Z
		endPos.Z=beginPos.Z
		beginPos.Z=temp
	}
	startZ,endZ=beginPos.Z,endPos.Z
	try_get:=C.phoenixbuilder_get_ranged_blocks(C.int(startX),C.int(beginPos.Y),C.int(startZ),C.int(endX),C.int(endPos.Y),C.int(endZ))
	palette:=make([]*string, 0)
	V:=(endPos.X-beginPos.X+1)*(endPos.Y-beginPos.Y+1)*(endPos.Z-beginPos.Z+1)
	blocks:=make([]*types.Module, V)
	if(int64(uintptr(unsafe.Pointer(try_get)))==0) {
		cmdsender.Output("Sorry, but this feature haven't implemented yet.")
		return nil
	}else if(int64(uintptr(unsafe.Pointer(try_get)))==3) {
		counter:=0
		for x:=beginPos.X; x<=endPos.X; x++ {
			for z:=beginPos.Z; z<=endPos.Z; z++ {
				for y:=beginPos.Y; y<=endPos.Y; y++ {
					blockCStr:=C.phoenixbuilder_get_block(C.int(x),C.int(y),C.int(z))
					blockStr:=C.GoString(blockCStr)
					C.free(unsafe.Pointer(blockCStr))
					var current *string=&blockStr
					for _, item := range palette {
						if(blockStr==*item) {
							current=item
							break
						}
					}
					if(&blockStr==current) {
						palette=append(palette, current)
					}
					blocks[counter]=&types.Module {
						Block: &types.Block {
							Name: current,
							Data: uint16(C.phoenixbuilder_get_block_data(C.int(x),C.int(y),C.int(z))),
						},
					}
					counter++
				}
			}
		}
		blocks=blocks[:counter]
	}else{
		counter:=0
		current_item:=""
		for {
			if(*((*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(try_get.palette))+uintptr(counter))))==0) {
				if(len(current_item)==0) {
					break
				}
				palette=append(palette, &current_item)
				if(counter==255) {
					panic("Too much items in palette!")
				}
				counter++
				continue
			}
			current_item+=string([]byte{byte(*((*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(try_get.palette))+uintptr(counter)))))})
			counter++
		}
		counter=0
		for x:=beginPos.X; x<=endPos.X; x++ {
			for z:=beginPos.Z; z<=endPos.Z; z++ {
				for y:=beginPos.Y; y<=endPos.Y; y++ {
					ref:=uint8(*((*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(try_get.data))+uintptr(counter)))))
					blocks[counter]=&types.Module {
						Block: &types.Block {
							Name: palette[ref],
							Data: uint16(*((*C.ushort)(unsafe.Pointer(uintptr(unsafe.Pointer(try_get.data))+uintptr(counter*2))))),
						},
					}
					counter++
				}
			}
		}
		blocks=blocks[counter:]
		C.free(unsafe.Pointer(try_get.palette))
		C.free(unsafe.Pointer(try_get.blocks_map))
		C.free(unsafe.Pointer(try_get.data))
		if(uintptr(unsafe.Pointer(try_get.nbt_area))!=uintptr(0)) {
			C.free(unsafe.Pointer(try_get.nbt_area))
		}
		C.free(unsafe.Pointer(try_get))
	}
	runtime.GC()
	out:=bdump.BDumpLegacy {
		Blocks: blocks,
	}
	if(strings.LastIndex(cfg.Path,".bdx")!=len(cfg.Path)-4||len(cfg.Path)<4) {
		cfg.Path+=".bdx"
	}
	cmdsender.Output("EXPORT >> Writing output file")
	err, signerr:=out.WriteToFile(cfg.Path, env.LocalCert, env.LocalKey)
	if(err!=nil){
		cmdsender.Output(fmt.Sprintf("EXPORT >> ERROR: Failed to export: %v",err))
		return nil
	}else if(signerr!=nil) {
		cmdsender.Output(fmt.Sprintf("EXPORT >> Note: The file is unsigned since the following error was trapped: %v",signerr))
	}else{
		cmdsender.Output(fmt.Sprintf("EXPORT >> File signed successfully"))
	}
	cmdsender.Output(fmt.Sprintf("EXPORT >> Successfully exported your structure to %v",cfg.Path))
	runtime.GC()
	return nil
}

func CreateLegacyExportTask(commandLine string, env *environment.PBEnvironment) *task.Task {
	return CreateExportTask(commandLine, env)
}

