package task

import (
	"phoenixbuilder/minecraft"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/fastbuilder/configuration"
	"phoenixbuilder/fastbuilder/parsing"
	"phoenixbuilder/fastbuilder/bdump"
	"fmt"
	"strings"
	"runtime"
	"phoenixbuilder/fastbuilder/world_provider"
	"phoenixbuilder/dragonfly/server/block/cube"
	"phoenixbuilder/dragonfly/server/world"
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
	if err!=nil {
		command.Tellraw(conn, fmt.Sprintf("Failed to parse command: %v",err))
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
		command.Tellraw(conn, "EXPORT >> World interaction interface is occupied, failing")
		return nil
	}
	world_provider.NewWorld(conn)
	go func() {
		command.Tellraw(conn, "EXPORT >> Exporting...")
		blocks:=make([]*types.RuntimeModule,0)
		for x:=beginPos.X; x<=endPos.X; x++ {
			for z:=beginPos.Z; z<=endPos.Z; z++ {
				for y:=beginPos.Y; y<=endPos.Y; y++ {
					blk:=world_provider.CurrentWorld.Block(cube.Pos{x,y,z})
					runtimeId:=world.LoadRuntimeID(blk)
					if runtimeId==world_provider.AirRuntimeId {
						continue
					}else if runtimeId==50000000 {
						continue
					}
					block, item:=blk.EncodeBlock()
					var cbdata *types.CommandBlockData = nil
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
					blocks=append(blocks,&types.RuntimeModule {
						BlockRuntimeId: runtimeId,
						CommandBlockData: cbdata,
						Point: types.Position {
							X: x,
							Y: y,
							Z: z,
						},
					})
				}
			}
		}
		world_provider.DestroyWorld()
		out:=bdump.BDump {
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
		runtime.GC()
	} ()
	return nil
}

