package task

import (
	"fmt"
	"github.com/google/uuid"
	"phoenixbuilder/dragonfly/server/block/cube"
	"phoenixbuilder/dragonfly/server/world"
	"phoenixbuilder/fastbuilder/bdump"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/world_provider"
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/environment"
	"runtime"
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

func CreateExportTask(commandLine string, env *environment.PBEnvironment) *Task {
	cmdsender:=env.CommandSender
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig(env).Main())
	if err!=nil {
		cmdsender.Tellraw(fmt.Sprintf("Failed to parse command: %v",err))
		return nil
	}
	beginPos := cfg.Position
	endPos := cfg.End
	if(endPos.X-beginPos.X<0) {
		temp:=endPos.X
		endPos.X=beginPos.X
		beginPos.X=temp
	}
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
	if(world_provider.CurrentWorld!=nil) {
		cmdsender.Tellraw("EXPORT >> World interaction interface is occupied, failing")
		return nil
	}
	world_provider.NewWorld(env)
	go func() {
		defer func() {
			r:=recover()
			if r!=nil{
				fmt.Println("go routine @ fastbuilder.task export crashed ",r)
			}
		}()
		cmdsender.Tellraw("EXPORT >> Exporting...")
		V:=(endPos.X-beginPos.X+1)*(endPos.Y-beginPos.Y+1)*(endPos.Z-beginPos.Z+1)
		blocks:=make([]*types.RuntimeModule,V)
		counter:=0
		for x:=beginPos.X; x<=endPos.X; x++ {
			for z:=beginPos.Z; z<=endPos.Z; z++ {
				for y:=beginPos.Y; y<=endPos.Y; y++ {
					blk:=world_provider.CurrentWorld.Block(cube.Pos{x,y,z})
					runtimeId:=world.LoadRuntimeID(blk)
					if runtimeId==world_provider.AirRuntimeId {
						continue
					}
					block, item:=blk.EncodeBlock()
					var cbdata *types.CommandBlockData = nil
					var chestData *types.ChestData = nil
					if(block=="chest"||strings.Contains(block,"shulker_box")) {
						content:=item["Items"].([]interface{})
						chest:=make(types.ChestData, len(content))
						for index, iface := range content {
							i:=iface.(map[string]interface{})
							name:=i["Name"].(string)
							count:=i["Count"].(uint8)
							damage:=i["Damage"].(int16)
							slot:=i["Slot"].(uint8)
							name_mcnk:=name[10:]
							chest[index]=types.ChestSlot {
								Name: name_mcnk,
								Count: count,
								Damage: uint16(int(damage)),
								Slot: slot,
							}
						}
						chestData=&chest
					}
					if strings.Contains(block,"command_block") {
						var mode uint32
						if(block=="command_block"){
							mode=packet.CommandBlockImpulse
						}else if(block=="repeating_command_block"){
							mode=packet.CommandBlockRepeat
						}else if(block=="chain_command_block"){
							mode=packet.CommandBlockChain
						}
						cmd:=item["Command"].(string)
						cusname:=item["CustomName"].(string)
						exeft:=item["ExecuteOnFirstTick"].(uint8)
						tickdelay:=item["TickDelay"].(int32)
						aut:=item["auto"].(uint8)
						trackoutput:=item["TrackOutput"].(uint8)
						lo:=item["LastOutput"].(string)
						//conditionalmode:=item["conditionalMode"].(uint8)
						data:=item["data"].(int32)
						var conb bool
						if (data>>3)&1 == 1 {
							conb=true
						}else{
							conb=false
						}
						var exeftb bool
						if exeft==0 {
							exeftb=true
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
					blocks[counter]=&types.RuntimeModule {
						BlockRuntimeId: runtimeId,
						CommandBlockData: cbdata,
						ChestData: chestData,
						Point: types.Position {
							X: x,
							Y: y,
							Z: z,
						},
					}
					counter++
				}
			}
		}
		world_provider.DestroyWorld()
		blocks=blocks[:counter]
		runtime.GC()
		out:=bdump.BDump {
			Blocks: blocks,
		}
		if(strings.LastIndex(cfg.Path,".bdx")!=len(cfg.Path)-4||len(cfg.Path)<4) {
			cfg.Path+=".bdx"
		}
		cmdsender.Tellraw("EXPORT >> Writing output file")
		err, signerr:=out.WriteToFile(cfg.Path, env.LocalCert, env.LocalKey)
		if(err!=nil){
			cmdsender.Tellraw(fmt.Sprintf("EXPORT >> ERROR: Failed to export: %v",err))
			return
		}else if(signerr!=nil) {
			cmdsender.Tellraw(fmt.Sprintf("EXPORT >> Note: The file is unsigned since the following error was trapped: %v",signerr))
		}else{
			cmdsender.Tellraw(fmt.Sprintf("EXPORT >> File signed successfully"))
		}
		cmdsender.Tellraw(fmt.Sprintf("EXPORT >> Successfully exported your structure to %v",cfg.Path))
		runtime.GC()
	} ()
	return nil
}

func CreateLegacyExportTask(commandLine string, env *environment.PBEnvironment) *Task {
	cfg, err := parsing.Parse(commandLine, configuration.GlobalFullConfig(env).Main())
	if err!=nil {
		env.CommandSender.Tellraw(fmt.Sprintf("Failed to parse command: %v", err))
		return nil
	}
	beginPos := cfg.Position
	endPos   := cfg.End
	msizex:=0
	msizey:=0
	msizez:=0
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
	gsizez:=msizez
	go func() {
		u_d, _ := uuid.NewUUID()
		env.CommandSender.SendWSCommand("gamemode c", u_d)
		originx:=0
		originz:=0
		var blocks []*types.Module
		for {
			env.CommandSender.Tellraw("EXPORT >> Fetching data")
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
			(*env.CommandSender.GetUUIDMap()).Store(u_d2.String(),wchan)
			env.CommandSender.SendWSCommand(fmt.Sprintf("tp %d %d %d",posx,beginPos.Y+1,posz), u_d2)
			<-wchan
			close(wchan)
			ExportWaiter=make(chan map[string]interface{})
			env.Connection.(*minecraft.Conn).WritePacket(&packet.StructureTemplateDataRequest {
				StructureName: "mystructure:a",
				Position: protocol.BlockPos {int32(posx),int32(beginPos.Y),int32(posz)},
				Settings: protocol.StructureSettings {
					PaletteName: "default",
					IgnoreEntities: true,
					IgnoreBlocks: false,
					Size: protocol.BlockPos {int32(cursizex),int32(msizey),int32(cursizez)},
					Offset: protocol.BlockPos {0,0,0},
					LastEditingPlayerUniqueID: env.Connection.(*minecraft.Conn).GameData().EntityUniqueID,
					Rotation: 0,
					Mirror: 0,
					Integrity: 100,
					Seed: 0,
				},
				RequestType: packet.StructureTemplateRequestExportFromSave,
			})
			exportData:=<-ExportWaiter
			close(ExportWaiter)
			env.CommandSender.Tellraw("EXPORT >> Data received, processing.")
			env.CommandSender.Tellraw("EXPORT >> Extracting blocks")
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
								Data:uint16(curblockdata),
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
		out:=bdump.BDumpLegacy {
			Blocks: blocks,
		}
		if(strings.LastIndex(cfg.Path,".bdx")!=len(cfg.Path)-4||len(cfg.Path)<4) {
			cfg.Path+=".bdx"
		}
		env.CommandSender.Tellraw("EXPORT >> Writing output file")
		err, signerr:=out.WriteToFile(cfg.Path, env.LocalCert, env.LocalKey)
		if(err!=nil){
			env.CommandSender.Tellraw(fmt.Sprintf("EXPORT >> ERROR: Failed to export: %v",err))
			return
		}else if(signerr!=nil) {
			env.CommandSender.Tellraw(fmt.Sprintf("EXPORT >> Note: The file is unsigned since the following error was trapped: %v",signerr))
		}else{
			env.CommandSender.Tellraw(fmt.Sprintf("EXPORT >> File signed successfully"))
		}
		env.CommandSender.Tellraw(fmt.Sprintf("EXPORT >> Successfully exported your structure to %v",cfg.Path))
	} ()
	return nil
}

