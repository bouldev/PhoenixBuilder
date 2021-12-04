package function

import (
	"os"
	"fmt"
	"phoenixbuilder/fastbuilder/types"
	"phoenixbuilder/fastbuilder/command"
	"phoenixbuilder/fastbuilder/configuration"
	fbtask "phoenixbuilder/fastbuilder/task"
	"phoenixbuilder/fastbuilder/builder"
	"phoenixbuilder/minecraft"
	"github.com/google/uuid"
	"phoenixbuilder/fastbuilder/i18n"
)



func InitInternalFunctions() {
	delayEnumId:=RegisterEnum("continuous, discrete, none",types.ParseDelayMode,types.DelayModeInvalid)
	RegisterFunction(&Function {
		Name: "exit",
		OwnedKeywords: []string {"fbexit"},
		FunctionType:FunctionTypeSimple,
		SFMinSliceLen: 1,
		FunctionContent: func(conn *minecraft.Conn,_ []interface{}) {
			command.Tellraw(conn,I18n.T(I18n.QuitCorrectly))
			fmt.Printf("%s\n",I18n.T(I18n.QuitCorrectly))
			conn.Close()
			os.Exit(0)
		},
	})
	RegisterFunction(&Function {
		Name: "ingameping",
		OwnedKeywords: []string {"ingameping"},
		FunctionType:FunctionTypeSimple,
		SFMinSliceLen: 1,
		FunctionContent: func(conn *minecraft.Conn,_ []interface{}) {
			command.SendSizukanaCommand("say Ingame pong",conn)
		},
	})
	RegisterFunction(&Function {
		Name: "set",
		OwnedKeywords: []string {"set"},
		FunctionType:FunctionTypeSimple,
		SFMinSliceLen:4,
		SFArgumentTypes: []byte {SimpleFunctionArgumentInt,SimpleFunctionArgumentInt,SimpleFunctionArgumentInt},
		FunctionContent: func(conn *minecraft.Conn,args []interface{}) {
			X, _ := args[0].(int)
			Y, _ := args[1].(int)
			Z, _ := args[2].(int)
			configuration.GlobalFullConfig().Main().Position=types.Position {
				X: X,
				Y: Y,
				Z: Z,
			}
			command.Tellraw(conn, fmt.Sprintf("%s: %d, %d, %d.",I18n.T(I18n.PositionSet),X,Y,Z))
		},
	})
	RegisterFunction(&Function {
		Name: "setend",
		OwnedKeywords: []string {"setend"},
		FunctionType:FunctionTypeSimple,
		SFMinSliceLen:4,
		SFArgumentTypes: []byte {SimpleFunctionArgumentInt,SimpleFunctionArgumentInt,SimpleFunctionArgumentInt},
		FunctionContent: func(conn *minecraft.Conn,args []interface{}) {
			X, _ := args[0].(int)
			Y, _ := args[1].(int)
			Z, _ := args[2].(int)
			configuration.GlobalFullConfig().Main().End=types.Position {
				X: X,
				Y: Y,
				Z: Z,
			}
			command.Tellraw(conn, fmt.Sprintf("%s: %d, %d, %d.",I18n.T(I18n.PositionSet_End),X,Y,Z))
		},
	})
	RegisterFunction(&Function {
		Name: "delay",
		OwnedKeywords: []string {"delay"},
		FunctionType: FunctionTypeContinue,
		SFMinSliceLen: 3,
		FunctionContent: map[string]*FunctionChainItem {
			"set": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn, args []interface{}){
					if configuration.GlobalFullConfig().Delay().DelayMode==types.DelayModeNone {
						command.Tellraw(conn, I18n.T(I18n.DelaySetUnavailableUnderNoneMode))
						return
					}
					ms, _:=args[0].(int)
					configuration.GlobalFullConfig().Delay().Delay=int64(ms)
					command.Tellraw(conn, fmt.Sprintf("%s: %d", I18n.T(I18n.DelaySet), ms))
				},
			},
			"mode": &FunctionChainItem {
				FunctionType: FunctionTypeContinue,
				Content: map[string]*FunctionChainItem {
					"get": &FunctionChainItem {
						FunctionType: FunctionTypeSimple,
						Content: func(conn *minecraft.Conn, _ []interface{}){
							command.Tellraw(conn, fmt.Sprintf("%s: %s.",I18n.T(I18n.CurrentDefaultDelayMode),types.StrDelayMode(configuration.GlobalFullConfig().Delay().DelayMode)))
						},
					},
					"set": &FunctionChainItem {
						FunctionType: FunctionTypeSimple,
						ArgumentTypes: []byte{byte(delayEnumId)},
						Content: func(conn *minecraft.Conn, args []interface{}){
							delaymode,_:=args[0].(byte)
							configuration.GlobalFullConfig().Delay().DelayMode=delaymode
							command.Tellraw(conn,fmt.Sprintf("%s: %s",I18n.T(I18n.DelayModeSet),types.StrDelayMode(delaymode)))
							if delaymode != types.DelayModeNone {
								dl:=decideDelay(delaymode)
								configuration.GlobalFullConfig().Delay().Delay=dl
								command.Tellraw(conn,fmt.Sprintf(I18n.T(I18n.DelayModeSet_DelayAuto),dl))
							}
							if delaymode==types.DelayModeDiscrete {
								configuration.GlobalFullConfig().Delay().DelayThreshold=decideDelayThreshold()
								command.Tellraw(conn,fmt.Sprintf(I18n.T(I18n.DelayModeSet_ThresholdAuto),configuration.GlobalFullConfig().Delay().DelayThreshold))
							}
						},
					},
				},
			},
			"threshold": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn, args []interface{}){
					if configuration.GlobalFullConfig().Delay().DelayMode != types.DelayModeDiscrete {
						command.Tellraw(conn, I18n.T(I18n.DelayThreshold_OnlyDiscrete))
						return
					}
					thr, _ := args[0].(int)
					configuration.GlobalFullConfig().Delay().DelayThreshold=thr
					command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.DelayThreshold_Set), thr))
				},
			},
		},
	})
	RegisterFunction(&Function {
		Name: "get-pos",
		OwnedKeywords: []string {"get"},
		FunctionType:FunctionTypeContinue,
		SFMinSliceLen:1,
		FunctionContent: map[string]*FunctionChainItem {
			"": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{},
				Content: func(conn *minecraft.Conn,_ []interface{}) {
					if(I18n.HasTranslationFor(I18n.Get_Warning)) {
						command.Tellraw(conn, I18n.T(I18n.Get_Warning))
					}
					command.SendCommand("gamerule sendcommandfeedback true",uuid.New(),conn)
					command.SendCommand(fmt.Sprintf("execute @a[name=\"%s\"] ~ ~ ~ testforblock ~ ~ ~ air",configuration.RespondUser),configuration.ZeroId,conn)
				},
			},
			"begin": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				Content: func(conn *minecraft.Conn,_ []interface{}) {
					if(I18n.HasTranslationFor(I18n.Get_Warning)) {
						command.Tellraw(conn, I18n.T(I18n.Get_Warning))
					}
					command.SendCommand("gamerule sendcommandfeedback true",uuid.New(),conn)
					command.SendCommand(fmt.Sprintf("execute @a[name=\"%s\"] ~ ~ ~ testforblock ~ ~ ~ air",configuration.RespondUser),configuration.ZeroId,conn)
				},
			},
			"end": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				Content: func(conn *minecraft.Conn,_ []interface{}) {
					if(I18n.HasTranslationFor(I18n.Get_Warning)) {
						command.Tellraw(conn, I18n.T(I18n.Get_Warning))
					}
					command.SendCommand("gamerule sendcommandfeedback true",uuid.New(),conn)
					command.SendCommand(fmt.Sprintf("execute @a[name=\"%s\"] ~ ~ ~ testforblock ~ ~ ~ air",configuration.RespondUser),configuration.OneId,conn)
				},
			},
		},
	})
	RegisterFunction(&Function {
		Name: "task",
		OwnedKeywords: []string {"task"},
		FunctionType: FunctionTypeContinue,
		SFMinSliceLen:2,
		FunctionContent: map[string]*FunctionChainItem {
			"list": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				Content: func(conn *minecraft.Conn, _ []interface{}){
					total:=0
					command.Tellraw(conn,I18n.T(I18n.CurrentTasks))
					fbtask.TaskMap.Range(func (_tid interface{}, _v interface{}) bool {
						tid,_:=_tid.(int64)
						v,_:=_v.(*fbtask.Task)
						dt:=-1
						dv:=int64(-1)
						if v.Config.Delay().DelayMode==types.DelayModeDiscrete {
							dt=v.Config.Delay().DelayThreshold
						}
						if v.Config.Delay().DelayMode!=types.DelayModeNone {
							dv=v.Config.Delay().Delay
						}
						command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.TaskStateLine),tid,v.CommandLine,fbtask.GetStateDesc(v.State),dv,types.StrDelayMode(v.Config.Delay().DelayMode),dt))
						total++
						return true
					})
					command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.TaskTotalCount),total))
				},
			},
			"pause": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn,args []interface{}) {
					tid, _ := args[0].(int)
					task:=fbtask.FindTask(int64(tid))
					if task==nil {
						command.Tellraw(conn, I18n.T(I18n.TaskNotFoundMessage))
						return
					}
					task.Pause()
					command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.TaskPausedNotice),task.TaskId))
				},
			},
			"resume": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn,args []interface{}) {
					tid, _ := args[0].(int)
					task:=fbtask.FindTask(int64(tid))
					if task==nil {
						command.Tellraw(conn, I18n.T(I18n.TaskNotFoundMessage))
						return
					}
					task.Resume()
					command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.TaskResumedNotice),task.TaskId))
				},
			},
			"break": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn,args []interface{}) {
					tid, _ := args[0].(int)
					task:=fbtask.FindTask(int64(tid))
					if task==nil {
						command.Tellraw(conn, I18n.T(I18n.TaskNotFoundMessage))
						return
					}
					task.Break()
					command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.TaskStoppedNotice),task.TaskId))
				},
			},
			"setdelay": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte {SimpleFunctionArgumentInt,SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn,args []interface{}) {
					tid, _ := args[0].(int)
					del, _ := args[1].(int)
					task:=fbtask.FindTask(int64(tid))
					if task==nil {
						command.Tellraw(conn, I18n.T(I18n.TaskNotFoundMessage))
						return
					}
					if(task.Config.Delay().DelayMode==types.DelayModeNone) {
						command.Tellraw(conn, I18n.T(I18n.Task_SetDelay_Unavailable))
						return
					}
					command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.Task_DelaySet),task.TaskId,del))
					task.Config.Delay().Delay=int64(del)
				},
			},
			"setdelaymode": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte {SimpleFunctionArgumentInt, byte(delayEnumId)},
				Content: func(conn *minecraft.Conn,args []interface{}) {
					tid, _ := args[0].(int)
					delaymode, _ := args[1].(byte)
					task:=fbtask.FindTask(int64(tid))
					if task==nil {
						command.Tellraw(conn, I18n.T(I18n.TaskNotFoundMessage))
						return
					}
					task.Pause()
					task.Config.Delay().DelayMode=delaymode
					command.Tellraw(conn, fmt.Sprintf("[%s %d] - %s: %s",I18n.T(I18n.TaskTTeIuKoto),tid,I18n.T(I18n.DelayModeSet),types.StrDelayMode(delaymode)))
					if delaymode!=types.DelayModeNone {
						task.Config.Delay().Delay=decideDelay(delaymode)
						command.Tellraw(conn, fmt.Sprintf("[%s %d] "+I18n.T(I18n.DelayModeSet_DelayAuto),I18n.T(I18n.TaskTTeIuKoto),task.TaskId,task.Config.Delay().Delay))
					}
					if delaymode==types.DelayModeDiscrete {
						task.Config.Delay().DelayThreshold=decideDelayThreshold()
						command.Tellraw(conn, fmt.Sprintf("[%s %d] "+I18n.T(I18n.DelayModeSet_ThresholdAuto),I18n.T(I18n.TaskTTeIuKoto),task.TaskId,task.Config.Delay().DelayThreshold))
					}
					task.Resume()
				},
			},
			"setdelaythreshold": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte {SimpleFunctionArgumentInt,SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn,args []interface{}) {
					tid, _ := args[0].(int)
					delayt, _ := args[1].(int)
					task:=fbtask.FindTask(int64(tid))
					if task==nil {
						command.Tellraw(conn, I18n.T(I18n.TaskNotFoundMessage))
						return
					}
					if task.Config.Delay().DelayMode!=types.DelayModeDiscrete {
						command.Tellraw(conn, I18n.T(I18n.DelayThreshold_OnlyDiscrete))
						return
					}
					command.Tellraw(conn, fmt.Sprintf("[%s %d] - "+I18n.T(I18n.DelayThreshold_Set),I18n.T(I18n.TaskTTeIuKoto),tid,delayt))
					task.Config.Delay().DelayThreshold=delayt
				},
			},
		},
	})
	taskTypeEnumId:=RegisterEnum("async, sync",types.ParseTaskType,types.TaskTypeInvalid)
	RegisterFunction(&Function {
		Name: "set task type",
		OwnedKeywords: []string{"tasktype"},
		FunctionType: FunctionTypeSimple,
		SFMinSliceLen: 2,
		SFArgumentTypes: []byte{byte(taskTypeEnumId)},
		FunctionContent: func(conn *minecraft.Conn,args []interface{}) {
			ev, _:=args[0].(byte)
			configuration.GlobalFullConfig().Global().TaskCreationType=ev
			command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.TaskTypeSwitchedTo),types.MakeTaskType(ev)))
		},
	})
	taskDMEnumId:=RegisterEnum("true, false",types.ParseTaskDisplayMode,types.TaskDisplayInvalid)
	RegisterFunction(&Function {
		Name: "set progress title display type",
		OwnedKeywords: []string{"progress"},
		FunctionType: FunctionTypeSimple,
		SFMinSliceLen: 2,
		SFArgumentTypes: []byte{byte(taskDMEnumId)},
		FunctionContent: func(conn *minecraft.Conn,args []interface{}) {
			ev, _:=args[0].(byte)
			configuration.GlobalFullConfig().Global().TaskDisplayMode=ev
			command.Tellraw(conn, fmt.Sprintf(I18n.T(I18n.TaskDisplayModeSet),types.MakeTaskDisplayMode(ev)))
		},
	})
	// ippan
	var builderMethods []string
	for met,_ := range builder.Builder {
		builderMethods=append(builderMethods,met)
	}
	RegisterFunction(&Function {
		Name: "ippanbrd",
		OwnedKeywords: builderMethods,
		FunctionType:FunctionTypeRegular,
		FunctionContent: func(conn *minecraft.Conn,msg string){
			task := fbtask.CreateTask(msg, conn)
			if task==nil {
				return
			}
			command.Tellraw(conn, fmt.Sprintf("%s, ID=%d.",I18n.T(I18n.TaskCreated),task.TaskId))
		},
	})
	RegisterFunction(&Function {
		Name: "export",
		OwnedKeywords: []string{"export"},
		FunctionType: FunctionTypeRegular,
		FunctionContent: func(conn *minecraft.Conn,msg string){
			task := fbtask.CreateExportTask(msg, conn)
			if task==nil {
				return
			}
			command.Tellraw(conn, fmt.Sprintf("%s, ID=%d.",I18n.T(I18n.TaskCreated),task.TaskId))
		},
	})
}


func decideDelay(delaytype byte) int64 {
	// Will add system check later,so don't merge into other functions.
	if delaytype==types.DelayModeContinuous {
		return 1000
	}else if delaytype==types.DelayModeDiscrete {
		return 15
	}else{
		return 0
	}
}

func decideDelayThreshold() int {
	// Will add system check later,so don't merge into other functions.
	return 20000
}
