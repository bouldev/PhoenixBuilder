package task

import (
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/bdump"
	"phoenixbuilder/minecraft/protocol"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"strconv"
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
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig().Main())
	//cfg.Execute = "export"
	if err!=nil {
		command.Tellraw(conn, fmt.Sprintf("Failed to parse command: %v",err))
		return nil
	}
	//fcfg := configuration.ConcatFullConfig(cfg, configuration.GlobalFullConfig().Delay())
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
	/*offsetx:=0
	offsety:=0
	offsetz:=0*/
	msizex:=0
	msizey:=0
	msizez:=0
	/*if(endPos.X-beginPos.X>=0) {
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
	}*/
	if(endPos.X-beginPos.X<0) {
		temp:=endPos.X
		endPos.X=beginPos.X
		beginPos.X=temp
	}
	msizex=endPos.X-beginPos.X+1
	if(endPos.Y-beginPos.Y<0) {
		temp:=endPos.Y
		endPos.Y=beginPos.Y
		beginPos.Y=temp
	}
	msizey=endPos.Y-beginPos.Y+1
	if(endPos.Z-beginPos.Z<0) {
		temp:=endPos.Z
		endPos.Z=beginPos.Z
		beginPos.Z=temp
	}
	msizez=endPos.Z-beginPos.Z+1
	//gsizex:=msizex
	gsizez:=msizez
	//fmt.Printf("%v,%v\n%v,%v,%v\n%v,%v,%v\n",beginPos,endPos,offsetx,offsety,offsetz,sizex,sizey,sizez)
	//return nil
	go func() {
		u_d, _ := uuid.NewUUID()
		command.SendWSCommand("gamemode c", u_d, conn)
		originx:=0
		originz:=0
		var blocks []*types.Module
		for {
			command.Tellraw(conn, "EXPORT >> Fetching data")
			cursizex:=msizex
			cursizez:=msizez
			if msizex>100 {
				cursizex=100
			}
			if msizez>100 {
				cursizez=100
			}
			posx:=beginPos.X+originx*100
			posz:=beginPos.Z+originz*100
			u_d2, _ := uuid.NewUUID()
			wchan:=make(chan *packet.CommandOutput)
			command.UUIDMap.Store(u_d2.String(),wchan)
			command.SendWSCommand(fmt.Sprintf("tp %d %d %d",posx,beginPos.Y+1,posz), u_d2, conn)
			<-wchan
			close(wchan)
			ExportWaiter=make(chan map[string]interface{})
			conn.WritePacket(&packet.StructureTemplateDataRequest {
				StructureName: "mystructure:a",
				Position: protocol.BlockPos {int32(posx),int32(beginPos.Y),int32(posz)},
				Settings: protocol.StructureSettings {
					PaletteName: "default",
					IgnoreEntities: true,
					IgnoreBlocks: false,
					Size: protocol.BlockPos {int32(cursizex),int32(msizey),int32(cursizez)},
					Offset: protocol.BlockPos {0,0,0},
					LastEditingPlayerUniqueID: conn.GameData().EntityUniqueID,
					Rotation: 0,
					Mirror: 0,
					Integrity: 100,
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
			if len(sizeoo)==0 {
				originz++
				msizez-=cursizez
				if(msizez<=0){
					msizez=gsizez
					originz=0
					originx++
					msizex-=cursizex
				}
				if(msizex<=0) {
					break
				}
				continue
			}
			sizea,_:=sizeoo[0].(int32)
			sizeb,_:=sizeoo[1].(int32)
			sizec,_:=sizeoo[2].(int32)
			size:=[]int{int(sizea),int(sizeb),int(sizec)}
			structure, _:=exportData["structure"].(map[string]interface{})
			indicesP, _:=structure["block_indices"].([]interface{})
			indices,_:=indicesP[0].([]interface{})
			if len(indicesP)!=2 {
				panic(fmt.Errorf("Unexcepted indices data: %v\n",indices))
			}
			{
				ind,_:=indices[0].(int32)
				if ind==-1 {
					indices,_=indicesP[1].([]interface{})
				}
				ind,_=indices[0].(int32)
				if ind==-1 {
					panic(fmt.Errorf("Exchanged but still -1: %v\n",indices))
				}
			}
			blockpalettepar,_:=structure["palette"].(map[string]interface{})
			blockpalettepar2,_:=blockpalettepar["default"].(map[string]interface{})
			blockpalette,_:=blockpalettepar2["block_palette"].([]/*map[string]*/interface{})
			blockposdata,_:=blockpalettepar2["block_position_data"].(map[string]interface{})
			airind:=int32(-1)
			i:=0
			for x:=0;x<size[0];x++ {
				for y:=0;y<size[1];y++ {
					for z:=0;z<size[2];z++ {
						ind,_:=indices[i].(int32)
						if ind==-1 {
							i++
							continue
						}
						if ind==airind {
							i++
							continue
						}
						curblock,_:=blockpalette[ind].(map[string]interface{})
						curblocknameunsplitted,_:=curblock["name"].(string)
						curblocknamesplitted:=strings.Split(curblocknameunsplitted,":")
						curblockname:=curblocknamesplitted[1]
						var cbdata *types.CommandBlockData=nil
						if curblockname=="air" {
							i++
							airind=ind
							continue
						}else if(!cfg.ExcludeCommands&&strings.Contains(curblockname,"command_block")) {
							itemp,_:=blockposdata[strconv.Itoa(i)].(map[string]interface{})
							item,_:=itemp["block_entity_data"].(map[string]interface{})
							var mode uint32
							if(curblockname=="command_block"){
								mode=packet.CommandBlockImpulse
							}else if(curblockname=="repeating_command_block"){
								mode=packet.CommandBlockRepeat
							}else if(curblockname=="chain_command_block"){
								mode=packet.CommandBlockChain
							}
							cmd,_:=item["Command"].(string)
							cusname,_:=item["CustomName"].(string)
							exeft,_:=item["ExecuteOnFirstTick"].(uint8)
							tickdelay,_:=item["TickDelay"].(int32)//*/
							aut,_:=item["auto"].(uint8)//!needrestone
							trackoutput,_:=item["TrackOutput"].(uint8)//
							lo,_:=item["LastOutput"].(string)
							conditionalmode:=item["conditionalMode"].(uint8)
							var exeftb bool
							if exeft==0 {
								exeftb=false
							}else{
								exeftb=true
							}
							var tob bool
							if trackoutput==1 {
								tob=true
							}else{
								tob=false
							}
							var nrb bool
							if aut==1 {
								nrb=false
								//REVERSED!!
							}else{
								nrb=true
							}
							var conb bool
							if conditionalmode==1 {
								conb=true
							}else{
								conb=false
							}
							cbdata=&types.CommandBlockData {
								Mode: mode,
								Command: cmd,
								CustomName: cusname,
								ExecuteOnFirstTick: exeftb,
								LastOutput: lo,
								TickDelay: tickdelay,
								TrackOutput: tob,
								Conditional: conb,
								NeedRedstone: nrb,
							}
						}
						curblockdata,_:=curblock["val"].(int16)
						blocks=append(blocks,&types.Module{
							Block: &types.Block {
								Name:&curblockname,
								Data:curblockdata,
							},
							CommandBlockData: cbdata,
							Point: types.Position {
								X: originx*100+x,
								Y: y,
								Z: originz*100+z,
							},
						})
						i++
					}
				}
			}
			originz++
			msizez-=cursizez
			if(msizez<=0){
				msizez=gsizez
				originz=0
				originx++
				msizex-=cursizex
			}
			if(msizex<=0) {
				break
			}
		}
		out:=bdump.BDump {
			Author: configuration.RespondUser,
			Blocks: blocks,
		}
		if(strings.LastIndex(cfg.Path,".bdx")!=len(cfg.Path)-4||len(cfg.Path)<4) {
			cfg.Path+=".bdx"
		}
		command.Tellraw(conn,"EXPORT >> Writing output file")
		err, signerr:=out.WriteToFile(cfg.Path)
		if(err!=nil){
			command.Tellraw(conn,fmt.Sprintf("EXPORT >> ERROR: Failed to export: %v",err))
			return
		}else if(signerr!=nil) {
			command.Tellraw(conn,fmt.Sprintf("EXPORT >> Note: The file is unsigned since the following error was trapped: %v",signerr))
		}else{
			command.Tellraw(conn,fmt.Sprintf("EXPORT >> File signed successfully"))
		}
		command.Tellraw(conn, fmt.Sprintf("EXPORT >> Successfully exported your structure to %v",cfg.Path))
	} ()
	return nil
}
