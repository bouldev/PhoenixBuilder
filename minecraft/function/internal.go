package function

import (
	"os"
	"fmt"
	"phoenixbuilder/minecraft/mctype"
	"phoenixbuilder/minecraft/command"
	"phoenixbuilder/minecraft/configuration"
	"phoenixbuilder/minecraft/fbtask"
	"phoenixbuilder/minecraft/builder"
	"phoenixbuilder/minecraft/enchant"
	"phoenixbuilder/minecraft"
	"github.com/google/uuid"
)



func InitInternalFunctions() {
	delayEnumId:=RegisterEnum("continuous, discrete, none",mctype.ParseDelayMode,mctype.DelayModeInvalid)
	RegisterFunction(&Function {
		Name: "exit",
		OwnedKeywords: []string {"fbexit"},
		FunctionType:FunctionTypeSimple,
		SFMinSliceLen: 1,
		FunctionContent: func(conn *minecraft.Conn,_ []interface{}) {
			command.Tellraw(conn,"Quit correctly")
			fmt.Printf("Quit correctly\n")
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
			configuration.GlobalFullConfig().Main().Position=mctype.Position {
				X: X,
				Y: Y,
				Z: Z,
			}
			command.Tellraw(conn, fmt.Sprintf("Position set: %d, %d, %d.",X,Y,Z))
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
			configuration.GlobalFullConfig().Main().End=mctype.Position {
				X: X,
				Y: Y,
				Z: Z,
			}
			command.Tellraw(conn, fmt.Sprintf("End position set: %d, %d, %d.",X,Y,Z))
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
					if configuration.GlobalFullConfig().Delay().DelayMode==mctype.DelayModeNone {
						command.Tellraw(conn, "[delay set] is unavailable with delay mode: none")
						return
					}
					ms, _:=args[0].(int)
					configuration.GlobalFullConfig().Delay().Delay=int64(ms)
					command.Tellraw(conn, fmt.Sprintf("Delay set: %d", ms))
				},
			},
			"mode": &FunctionChainItem {
				FunctionType: FunctionTypeContinue,
				Content: map[string]*FunctionChainItem {
					"get": &FunctionChainItem {
						FunctionType: FunctionTypeSimple,
						Content: func(conn *minecraft.Conn, _ []interface{}){
							command.Tellraw(conn, fmt.Sprintf("Current default delay mode: %s.",mctype.StrDelayMode(configuration.GlobalFullConfig().Delay().DelayMode)))
						},
					},
					"set": &FunctionChainItem {
						FunctionType: FunctionTypeSimple,
						ArgumentTypes: []byte{byte(delayEnumId)},
						Content: func(conn *minecraft.Conn, args []interface{}){
							delaymode,_:=args[0].(byte)
							configuration.GlobalFullConfig().Delay().DelayMode=delaymode
							command.Tellraw(conn,fmt.Sprintf("Delay mode set: %s",mctype.StrDelayMode(delaymode)))
							if delaymode != mctype.DelayModeNone {
								dl:=decideDelay(delaymode)
								configuration.GlobalFullConfig().Delay().Delay=dl
								command.Tellraw(conn,fmt.Sprintf("Delay automatically set to: %d",dl))
							}
							if delaymode==mctype.DelayModeDiscrete {
								configuration.GlobalFullConfig().Delay().DelayThreshold=decideDelayThreshold()
								command.Tellraw(conn,fmt.Sprintf("Delay threshold automatically set to: %d",configuration.GlobalFullConfig().Delay().DelayThreshold))
							}
						},
					},
				},
			},
			"threshold": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn, args []interface{}){
					if configuration.GlobalFullConfig().Delay().DelayMode != mctype.DelayModeDiscrete {
						command.Tellraw(conn, "Delay threshold is only available with delay mode: discrete")
						return
					}
					thr, _ := args[0].(int)
					configuration.GlobalFullConfig().Delay().DelayThreshold=thr
					command.Tellraw(conn, fmt.Sprintf("Delay threshold set to: %d", thr))
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
					command.SendCommand("gamerule sendcommandfeedback true",uuid.New(),conn)
					command.SendCommand(fmt.Sprintf("execute @a[name=\"%s\"] ~ ~ ~ testforblock ~ ~ ~ air",configuration.RespondUser),configuration.ZeroId,conn)
				},
			},
			"begin": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				Content: func(conn *minecraft.Conn,_ []interface{}) {
					command.SendCommand("gamerule sendcommandfeedback true",uuid.New(),conn)
					command.SendCommand(fmt.Sprintf("execute @a[name=\"%s\"] ~ ~ ~ testforblock ~ ~ ~ air",configuration.RespondUser),configuration.ZeroId,conn)
				},
			},
			"end": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				Content: func(conn *minecraft.Conn,_ []interface{}) {
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
					command.Tellraw(conn,"Current tasks:")
					fbtask.TaskMap.Range(func (_tid interface{}, _v interface{}) bool {
						tid,_:=_tid.(int64)
						v,_:=_v.(*fbtask.Task)
						dt:=-1
						dv:=int64(-1)
						if v.Config.Delay().DelayMode==mctype.DelayModeDiscrete {
							dt=v.Config.Delay().DelayThreshold
						}
						if v.Config.Delay().DelayMode!=mctype.DelayModeNone {
							dv=v.Config.Delay().Delay
						}
						command.Tellraw(conn, fmt.Sprintf("ID %d - CommandLine:\"%s\", State: %s, Delay: %d, DelayMode: %s, DelayThreshold: %d",tid,v.CommandLine,fbtask.GetStateDesc(v.State),dv,mctype.StrDelayMode(v.Config.Delay().DelayMode),dt))
						total++
						return true
					})
					command.Tellraw(conn, fmt.Sprintf("Total: %d",total))
				},
			},
			"pause": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn,args []interface{}) {
					tid, _ := args[0].(int)
					task:=fbtask.FindTask(int64(tid))
					if task==nil {
						command.Tellraw(conn, "Couldn't find a valid task by provided task id.")
						return
					}
					task.Pause()
					command.Tellraw(conn, fmt.Sprintf("[Task %d] - Paused",task.TaskId))
				},
			},
			"resume": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn,args []interface{}) {
					tid, _ := args[0].(int)
					task:=fbtask.FindTask(int64(tid))
					if task==nil {
						command.Tellraw(conn, "Couldn't find a valid task by provided task id.")
						return
					}
					task.Resume()
					command.Tellraw(conn, fmt.Sprintf("[Task %d] - Resumed",task.TaskId))
				},
			},
			"break": &FunctionChainItem {
				FunctionType: FunctionTypeSimple,
				ArgumentTypes: []byte{SimpleFunctionArgumentInt},
				Content: func(conn *minecraft.Conn,args []interface{}) {
					tid, _ := args[0].(int)
					task:=fbtask.FindTask(int64(tid))
					if task==nil {
						command.Tellraw(conn, "Couldn't find a valid task by provided task id.")
						return
					}
					task.Break()
					command.Tellraw(conn, fmt.Sprintf("[Task %d] - Stopped",task.TaskId))
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
						command.Tellraw(conn, "Couldn't find a valid task by provided task id.")
						return
					}
					if(task.Config.Delay().DelayMode==mctype.DelayModeNone) {
						command.Tellraw(conn, "[setdelay] is unavailable with delay mode: none")
						return
					}
					command.Tellraw(conn, fmt.Sprintf("[Task %d] - Delay set: %d",task.TaskId,del))
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
						command.Tellraw(conn, "Couldn't find a valid task by provided task id.")
						return
					}
					task.Pause()
					task.Config.Delay().DelayMode=delaymode
					command.Tellraw(conn, fmt.Sprintf("[Task %d] - Delay mode set: %s",tid,mctype.StrDelayMode(delaymode)))
					if delaymode!=mctype.DelayModeNone {
						task.Config.Delay().Delay=decideDelay(delaymode)
						command.Tellraw(conn, fmt.Sprintf("[Task %d] Delay automatically set to: %d",task.TaskId,task.Config.Delay().Delay))
					}
					if delaymode==mctype.DelayModeDiscrete {
						task.Config.Delay().DelayThreshold=decideDelayThreshold()
						command.Tellraw(conn, fmt.Sprintf("[Task %d] Delay threshold automatically set to: %d",task.TaskId,task.Config.Delay().DelayThreshold))
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
						command.Tellraw(conn, "Couldn't find a valid task by provided task id.")
						return
					}
					if task.Config.Delay().DelayMode!=mctype.DelayModeDiscrete {
						command.Tellraw(conn, "Delay threshold is only available with delay mode: discrete.")
						return
					}
					command.Tellraw(conn, fmt.Sprintf("[Task %d] - Delay threshold set: %d",tid,delayt))
					task.Config.Delay().DelayThreshold=delayt
				},
			},
		},
	})
	taskTypeEnumId:=RegisterEnum("async, sync",mctype.ParseTaskType,mctype.TaskTypeInvalid)
	RegisterFunction(&Function {
		Name: "set task type",
		OwnedKeywords: []string{"tasktype"},
		FunctionType: FunctionTypeSimple,
		SFMinSliceLen: 2,
		SFArgumentTypes: []byte{byte(taskTypeEnumId)},
		FunctionContent: func(conn *minecraft.Conn,args []interface{}) {
			ev, _:=args[0].(byte)
			configuration.GlobalFullConfig().Global().TaskCreationType=ev
			command.Tellraw(conn, fmt.Sprintf("Task creation type set to: %s.",mctype.MakeTaskType(ev)))
		},
	})
	taskDMEnumId:=RegisterEnum("true, false",mctype.ParseTaskDisplayMode,mctype.TaskDisplayInvalid)
	RegisterFunction(&Function {
		Name: "set progress title display type",
		OwnedKeywords: []string{"progress"},
		FunctionType: FunctionTypeSimple,
		SFMinSliceLen: 2,
		SFArgumentTypes: []byte{byte(taskDMEnumId)},
		FunctionContent: func(conn *minecraft.Conn,args []interface{}) {
			ev, _:=args[0].(byte)
			configuration.GlobalFullConfig().Global().TaskDisplayMode=ev
			command.Tellraw(conn, fmt.Sprintf("Task status display mode set to: %s.",mctype.MakeTaskDisplayMode(ev)))
		},
	})
	RegisterFunction(&Function {
		Name: "enchant",
		OwnedKeywords: []string {"enchant"},
		FunctionType: FunctionTypeSimple,
		SFMinSliceLen: 1,
		FunctionContent: func(conn *minecraft.Conn,args []interface{}) {
			enchant.Run(conn)
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
			command.Tellraw(conn, fmt.Sprintf("Task Created, ID=%d.",task.TaskId))
		},
	})
	RegisterFunction(&Function {
		Name: "export",
		OwnedKeywords: []string{"export"},
		FunctionType: FunctionTypeRegular,
		FunctionContent: func(conn *minecraft.Conn,msg string){
			//command.Tellraw(conn, "Unpublished function")
			//return
			task := fbtask.CreateExportTask(msg, conn)
			if task==nil {
				return
			}
			command.Tellraw(conn, fmt.Sprintf("Task Created, ID=%d.",task.TaskId))
		},
	})
}


func decideDelay(delaytype byte) int64 {
	// Will add system check later,so don't merge into other functions.
	if delaytype==mctype.DelayModeContinuous {
		return 1000
	}else if delaytype==mctype.DelayModeDiscrete {
		return 15
	}else{
		return 0
	}
}

func decideDelayThreshold() int {
	// Will add system check later,so don't merge into other functions.
	return 20000
}
