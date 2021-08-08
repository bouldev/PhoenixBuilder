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
	conn.WritePacket(&packet.BlockPickRequest {
		Position: protocol.BlockPos {int32(beginPos.X),int32(beginPos.Y),int32(beginPos.Z)},
		AddBlockNBT: false,
		HotBarSlot: 0,
	})
	return nil
	if beginPos.X > endPos.X {
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
	}
	totalT:=0
	for x:=beginPos.X;x<=endPos.X;x++ {
		for y:=beginPos.Y;y<=endPos.Y;y++ {
			for z:=beginPos.Z;z<=endPos.Z;z++ {
				totalT++
			}
		}
	}
	task := &Task {
		TaskId: TaskIdCounter.Add(1),
		CommandLine: commandLine,
		OutputChannel: nil,
		State: TaskStateRunning,
		Type: mctype.TaskTypeAsync,
		AsyncInfo: AsyncInfo {
			Built: 0,
			Total: totalT,
		},
		Config: fcfg,
	}
	var BuiltNumMutex sync.Mutex
	taskid := task.TaskId
	TaskMap.Store(taskid, task)
	go func(){
		threadNum:=atomic.NewInt64(0)
		threadDieChan:=make(chan bool)//threadDieちゃん（
		for x:=beginPos.X;x<=endPos.X;x++ {
			for z:=beginPos.Z;z<=endPos.Z;z++ {
				if threadNum.Load()>16 {
					<-threadDieChan
				}
				threadNum.Add(1)
				go func() {
					cmdreceiverchan:=make(chan *packet.CommandOutput)
					cud,_:=uuid.NewUUID()
					command.UUIDMap.Store(cud.String(),cmdreceiverchan)
					top:=endPos.Y
					command.SendWSCommand(fmt.Sprintf("gettopsolidblock %d %d %d",x,top,z),cud,conn)
					cmdcontent:=<-cmdreceiverchan
					if len(cmdcontent.OutputMessages)!=0 {
						jsonStr:=[]byte(cmdcontent.UnknownString)
						var ret SolidRet
						if err:=json.Unmarshal(jsonStr,&ret); err!=nil {
							panic(fmt.Errorf("Failed to parse[ex]: %v",err))
						}
						top=int(ret.Position.Y)
						bbt:=endPos.Y-top
						BuiltNumMutex.Lock()
						task.AsyncInfo.Built+=bbt
						BuiltNumMutex.Unlock()
					}
					for y:=beginPos.Y;y<=top;y++ {
						command.UUIDMap.Store(cud.String(),cmdreceiverchan)
						command.SendWSCommand(fmt.Sprintf("testforblock %d %d %d air",x,y,z),cud,conn)
						cmdcontent:=<-cmdreceiverchan
						if(cmdcontent.SuccessCount!=0){
							// is air
							continue
						}
						
					}
					threadNum.Add(-1)
					select {
					case threadDieChan<-true:
						return
					default:
						return
					}
				}()
			}
		}
		/*cmdreceiverchan:=make(chan *packet.CommandOutput,10240)
		cud, _ := uuid.NewUUID()
		command.UUIDMap.Store(cud.String(),cmdreceiverchan)
		command.SendSizukanaCommand(fmt.Sprintf("gettopsolidblock %d %d %d",endPos.X,endPos.Y,endPos.Z),cud,conn)
		fmt.Printf("%+v\n",<-cmdreceiverchan)
		close(cmdreceiverchan)*/
	}()
	return task
}