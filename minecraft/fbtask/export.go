package fbtask

import (
	"phoenixbuilder/minecraft"
	//"phoenixbuilder/minecraft/hotbarmanager"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/minecraft/mctype"
	"phoenixbuilder/minecraft/command"
	"phoenixbuilder/minecraft/configuration"
	"fmt"
	"github.com/google/uuid"
	"phoenixbuilder/minecraft/parse"
	"phoenixbuilder/minecraft/protocol"
	"encoding/json"
	"go.uber.org/atomic"
	"sync"
	"strings"
	"phoenixbuilder/minecraft/bdump"
)


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

func CreateExportTask(commandLine string, conn *minecraft.Conn) *Task {
	cfg, err := parse.Parse(commandLine, configuration.GlobalFullConfig().Main())
	//cfg.Execute = "export"
	if err!=nil {
		command.Tellraw(conn, fmt.Sprintf("Failed to parse command: %v",err))
		return nil
	}
	fcfg := configuration.ConcatFullConfig(cfg, configuration.GlobalFullConfig().Delay())
	//dcfg := fcfg.Delay()
	beginPos := cfg.Position
	endPos   := cfg.End
	/*conn.WritePacket(&packet.BlockPickRequest {
		Position: protocol.BlockPos {int32(beginPos.X),int32(beginPos.Y),int32(beginPos.Z)},
		AddBlockNBT: true,
		HotBarSlot: 0,
	})*/
	/*if beginPos.X > endPos.X {
		f:=beginPos.X
		endPos.X=beginPos.X
		beginPos.X=f
	}
	if beginPos.Y > endPos.Y {
		f:=beginPos.Y
		endPos.Y=beginPos.Y
		beginPos.Y=f
	}
	if beginPos.Z > endPos.Z {
		f:=beginPos.Z
		endPos.Z=beginPos.Z
		beginPos.Z=f
	}*/
	offsetx:=0
	offsety:=0
	offsetz:=0
	sizex:=0
	sizey:=0
	sizez:=0
	if(endPos.X-beginPos.X>=0) {
		sizex=endPos.X-beginPos.X
	}else{
		offsetx=endPos.X-beginPos.X
		sizex=beginPos.X-endPos.X
	}
	if(endPos.Y-beginPos.Y>=0) {
		sizey=endPos.Y-beginPos.Y
	}else{
		offsety=endPos.Y-beginPos.Y
		sizey=beginPos.Y-endPos.Y
	}
	if(endPos.Z-beginPos.Z>=0) {
		sizez=endPos.Z-beginPos.Z
	}else{
		offsetz=endPos.Z-beginPos.Z
		sizez=beginPos.Z-endPos.Z
	}
	//fmt.Printf("%v,%v\n%v,%v,%v\n%v,%v,%v\n",beginPos,endPos,offsetx,offsety,offsetz,sizex,sizey,sizez)
	//return nil
	go func() {
		u_d, _ := uuid.NewUUID()
		u_d2, _ := uuid.NewUUID()
		command.SendWSCommand("gamemode c", u_d, conn)
		command.SendWSCommand(fmt.Sprintf("tp %d %d %d",beginPos.X,beginPos.Y+1,beginPos.Z), u_d2, conn)
		command.Tellraw(conn, "EXPORT >> Fetching data")
		ExportWaiter=make(chan map[string]interface{})
		conn.WritePacket(&packet.StructureTemplateDataRequest {
			StructureName: "mystructure:a",
			Position: protocol.BlockPos {int32(beginPos.X),int32(beginPos.Y),int32(beginPos.Z)},
			Settings: protocol.StructureSettings {
				PaletteName: "default",
				IgnoreEntities: true,
				IgnoreBlocks: false,
				Size: protocol.BlockPos {int32(sizex),int32(sizey),int32(sizez)},
				Offset: protocol.BlockPos {int32(offsetx),int32(offsety),int32(offsetz)},
				LastEditingPlayerUniqueID: conn.GameData().EntityUniqueID,
				Rotation: 0,
				Mirror: 0,
				Integrity: 1,
				Seed: 0,
			},
			RequestType: packet.StructureTemplateRequestExportFromSave,
		})
		exportData:=<-ExportWaiter
		close(ExportWaiter)
		//fmt.Printf("%v",exportData["size"])
		command.Tellraw(conn, "EXPORT >> Data received, processing.")
		command.Tellraw(conn, "EXPORT >> Extracting blocks")
		sizeoo, _:=exportData["size"].([]interface{})
		sizex,_:=sizeoo[0].(int32)
		sizey,_:=sizeoo[1].(int32)
		sizez,_:=sizeoo[2].(int32)
		size:=[]int{int(sizex),int(sizey),int(sizez)}
		structure, _:=exportData["structure"].(map[string]interface{})
		indicesP, _:=structure["block_indices"].([]interface{})
		indices,_:=indicesP[0].([]interface{})
		blockpalettepar,_:=structure["palette"].(map[string]interface{})
		blockpalettepar2,_:=blockpalettepar["default"].(map[string]interface{})
		blockpalette,_:=blockpalettepar2["block_palette"].([]/*map[string]*/interface{})
		var blocks []*mctype.Module
		airind:=int32(-1)
		i:=0
		for x:=0;x<size[0];x++ {
			for y:=0;y<size[1];y++ {
				for z:=0;z<size[2];z++ {
					ind,_:=indices[i].(int32)
					if ind==airind {
						i++
						continue
					}
					curblock,_:=blockpalette[ind].(map[string]interface{})
					curblocknameunsplitted,_:=curblock["name"].(string)
					curblocknamesplitted:=strings.Split(curblocknameunsplitted,":")
					curblockname:=curblocknamesplitted[1]
					if curblockname=="air" {
						i++
						airind=ind
						continue
					}
					curblockdata,_:=curblock["val"].(int16)
					blocks=append(blocks,&mctype.Module{
						Block: &mctype.Block {
							Name:&curblockname,
							Data:curblockdata,
						},
						Point: mctype.Position {
							X: x,
							Y: y,
							Z: z,
						},
					})
					i++
				}
			}
		}
		out:=bdump.BDump {
			Author: configuration.RespondUser,
			Blocks: blocks,
		}
		if(strings.LastIndex(cfg.Path,".bdx")!=len(cfg.Path)-4) {
			cfg.Path+=".bdx"
		}
		command.Tellraw(conn,"EXPORT >> Writing output file")
		err:=out.WriteToFile(cfg.Path)
		if err!=nil {
			command.Tellraw(conn,fmt.Sprintf("EXPORT >> ERROR: Failed to export: %v",err))
			return
		}
		command.Tellraw(conn, fmt.Sprintf("EXPORT >> Successfully exported your structure to %v",cfg.Path))
	} ()
	return nil
}
