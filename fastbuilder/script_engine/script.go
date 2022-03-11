package script_engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"phoenixbuilder/fastbuilder/script_engine/built_in"
	"phoenixbuilder/minecraft/protocol/packet"
	"strings"

	"github.com/gorilla/websocket"
	"go.kuoruan.net/v8go-polyfills/base64"
	"go.kuoruan.net/v8go-polyfills/fetch"
	"go.kuoruan.net/v8go-polyfills/timers"
	"go.kuoruan.net/v8go-polyfills/url"
	"phoenixbuilder/fastbuilder/script_engine/bridge"
	"rogchap.com/v8go"
)

// jsEngine.hostBridge.api
const JSVERSION = "script_engine@v8.gamma.5"

func AllowPath(path string) bool {
	if strings.Contains(path, "fbtoken") {
		return false
	}
	if strings.Contains(path, "fb_script_permission") {
		return false
	}
	return true
}

func LoadPermission(hb bridge.HostBridge, identifyStr string) map[string]bool {
	permission := map[string]bool{}
	fullPermission := map[string]map[string]bool{}
	file, err := hb.LoadFile("fb_script_permission.json")
	if err != nil {
		return permission
	}
	err = json.Unmarshal([]byte(file), &fullPermission)
	if err != nil {
		return permission
	}
	if savedPermission, ok := fullPermission[identifyStr]; ok {
		return savedPermission
	}
	return permission
}

func SavePermission(hb bridge.HostBridge, identifyStr string, permission map[string]bool) {
	fullPermission := map[string]map[string]bool{}
	file, err := hb.LoadFile("fb_script_permission.json")
	dataToSave := []byte{}
	if err == nil {
		json.Unmarshal([]byte(file), &fullPermission)
	}
	fullPermission[identifyStr] = permission
	dataToSave, _ = json.Marshal(fullPermission)
	hb.SaveFile("fb_script_permission.json", string(dataToSave))
}

func getReceiver(info *v8go.FunctionCallbackInfo) *v8go.Object {
	return info.This()
}

func InitHostFns(iso *v8go.Isolate, global *v8go.ObjectTemplate, hb bridge.HostBridge, _scriptName string, identifyStr string, scriptPath string) func() {
	scriptName := _scriptName
	permission := LoadPermission(hb, identifyStr)
	updatePermission := func() {
		SavePermission(hb, identifyStr, permission)
	}

	throwException := func(funcName string, str string) *v8go.Value {
		value, _ := v8go.NewValue(iso, "Script crashed at ["+funcName+"] due to "+str)
		iso.ThrowException(value)
		return nil
	}
	printException := func(funcName string, str string) *v8go.Value {
		fmt.Println("Script triggered an exception at [" + funcName + "] due to " + str)
		return nil
	}
	throwNotConnectedException := func(funcName string) *v8go.Value {
		return throwException(funcName, "connection to MC not established")
	}
	hasStrIn := func(info *v8go.FunctionCallbackInfo, pos int, argName string) (string, bool) {
		if len(info.Args()) < pos+1 {
			return fmt.Sprintf("no arg %v provided in pos %v", argName, pos), false
		}
		if !info.Args()[pos].IsString() {
			return fmt.Sprintf("arg %v in pos %v is not a string (you set: %v)", argName, pos, info.Args()[pos].String()), false
		}
		return info.Args()[pos].String(), true
	}
	hasFuncIn := func(info *v8go.FunctionCallbackInfo, pos int, argName string) (string, *v8go.Function) {
		if len(info.Args()) < pos+1 {
			return fmt.Sprintf("no arg %v provided in pos %v", argName, pos), nil
		}
		function, err := info.Args()[pos].AsFunction()
		if err != nil {
			return fmt.Sprintf("arg %v in pos %v is not a function (you set: %v)", argName, pos, info.Args()[pos].String()), nil
		}
		return "", function
	}
	t := bridge.NewTerminator()
	t.TerminateHook = append(t.TerminateHook, func() {
		iso.TerminateExecution()
	})
	engine := v8go.NewObjectTemplate(iso)
	global.Set("engine", engine)
	// function engine.setName(scriptName string)
	if err := engine.Set("setName",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if str, ok := hasStrIn(info, 0, "engine.setName[scriptName]"); !ok {
				throwException("engine.setName: No arguments assigned", str)
			} else {
				hb.Println("Script \""+scriptName+"\" is naming itself as \""+str+"\"", t, scriptName)
				scriptName = str
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function engine.waitConnectionSync()
	if err := engine.Set("waitConnectionSync",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			hb.WaitConnect(t)
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function engine.waitConnection(cb)
	if err := engine.Set("waitConnection",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			_args := info.Args()
			if len(_args) == 0 {
				throwException("engine.waitConnection(cb)", " No arguments assigned")
			}
			first_arg := _args[0]
			if !first_arg.IsFunction() {
				throwException("engine.waitConnection(cb)", " Callback should be a function")
			}
			f, e := first_arg.AsFunction()
			if e != nil {
				throwException("engine.waitConnection(cb)", " Callback should be a function, but got function.")
			}
			go func() {
				hb.WaitConnect(t)
				f.Call(getReceiver(info))
			}()
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function engine.message(msg string) None
	if err := engine.Set("message",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if str, ok := hasStrIn(info, 0, "engine.message[msg]"); !ok {
				throwException("engine.message", str)
			} else {
				hb.Println(str, t, scriptName)
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function engine.questionSync(hint string) string
	if err := engine.Set("questionSync",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if str, ok := hasStrIn(info, 0, "engine.questionSync[hint]"); !ok {
				throwException("engine.questionSync", str)
			} else {
				userInput := hb.GetInput(str, t, scriptName)
				value, _ := v8go.NewValue(iso, userInput)
				return value
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function question(hint,cb) None
	if err := engine.Set("question",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if str, ok := hasStrIn(info, 0, "engine.question[hint,cb]"); !ok {
				throwException("engine.question", str)
			} else {
				if errStr, cbFn := hasFuncIn(info, 1, "engine.question[hint,cb]"); cbFn == nil {
					throwException("engine.question", errStr)
				} else {
					go func() {
						userInput := hb.GetInput(str, t, scriptName)
						value, _ := v8go.NewValue(iso, userInput)
						cbFn.Call(getReceiver(info), value)
					}()
				}

			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function engine.crash(string reason) None
	if err := engine.Set("crash",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {

			if str, ok := hasStrIn(info, 0, "engine.crash[reason]"); !ok {
				throwException("engine.crash", str)
			} else {
				throwException("engine.crash", str)
				t.Terminate()
			}
			return nil
		})); err != nil {
		panic(err)
	}

	// function fs.autoRestart() None
	// 这里做了指数退避，几次重连失败就会放缓到1小时重连一次，见 main.go 170行
	if err := engine.Set("autoRestart",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			hb.RequireAutoRestart()
			return nil
		})); err != nil {
		panic(err)
	}

	game := v8go.NewObjectTemplate(iso)
	global.Set("game", game)
	// One shot command
	// function game.eval(fbCmd string) None
	if err := game.Set("eval",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if !hb.IsConnected() {
				throwNotConnectedException("game.eval")
			}
			if str, ok := hasStrIn(info, 0, "game.eval[fbCmd]"); !ok {
				throwException("game.eval", str)
			} else {
				hb.FBCmd(str, t)
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function game.oneShotCommand(mcCmd string) None
	if err := game.Set("oneShotCommand",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if !hb.IsConnected() {
				throwNotConnectedException("game.oneShotCommand")
			}
			if str, ok := hasStrIn(info, 0, "game.oneShotCommand[mcCmd]"); !ok {
				throwException("game.oneShotCommand", str)
			} else {
				hb.MCCmd(str, t, false)
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function game.sendCommandSync(mcCmd string) jsObject
	if err := game.Set("sendCommandSync",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if !hb.IsConnected() {
				throwNotConnectedException("game.sendCommandSync")
			}
			if str, ok := hasStrIn(info, 0, "game.sendCommandSync[mcCmd]"); !ok {
				throwException("game.sendCommandSync", str)
			} else {
				pk := hb.MCCmd(str, t, true)
				strPk, err := json.Marshal(pk)
				if err != nil {
					return throwException("game.sendCommandSync", "Cannot convert host packet to Json Str: "+str)
				}
				value, err := v8go.JSONParse(info.Context(), string(strPk))
				if err != nil {
					return throwException("game.sendCommandSync", str)
				} else {
					return value
				}
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function game.sendCommand(mcCmd string, onResult function(jsObject)) None
	// jsObject=null, if cannot get result in callback
	if err := game.Set("sendCommand",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if !hb.IsConnected() {
				throwNotConnectedException("game.sendCommand")
			}
			if str, ok := hasStrIn(info, 0, "game.sendCommand[mcCmd]"); !ok {
				throwException("game.sendCommand", str)
			} else {
				if _, cbFn := hasFuncIn(info, 1, "game.sendCommand[onResult]"); cbFn == nil {
					hb.MCCmd(str, t, false)
					return nil
				} else {
					ctx := info.Context()
					go func() {
						pk := hb.MCCmd(str, t, true)
						strPk, err := json.Marshal(pk)
						if err != nil {
							printException("game.sendCommand", "Cannot convert host packet to Json Str: "+str)
							cbFn.Call(getReceiver(info), v8go.Null(iso))
							return
						}
						val, err := v8go.JSONParse(ctx, string(strPk))
						if err != nil {
							printException("game.sendCommand", "Cannot Parse Json Packet in Host: "+str)
							cbFn.Call(getReceiver(info), v8go.Null(iso))
							return
						} else {
							cbFn.Call(getReceiver(info), val)
						}
					}()
				}
				return nil
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	if err := game.Set("botPos",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			ot := v8go.NewObjectTemplate(iso)
			x, y, z := hb.GetBotPos()
			jsX, _ := v8go.NewValue(iso, int32(x))
			jsY, _ := v8go.NewValue(iso, int32(y))
			jsZ, _ := v8go.NewValue(iso, int32(z))
			ot.Set("x", jsX)
			ot.Set("y", jsY)
			ot.Set("z", jsZ)
			jsPos, _ := ot.NewInstance(info.Context())
			return jsPos.Value
		})); err != nil {
		panic(err)
	}

	// function game.subscribePacket(packetType,onPacketCb) deRegFn
	// when deRegFn is called, onPacketCb function will no longer be called
	if err := game.Set("subscribePacket",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if str, ok := hasStrIn(info, 0, "game.subscribePacket[packetType]"); !ok {
				throwException("game.subscribePacket", str)
			} else {
				if errStr, cbFn := hasFuncIn(info, 1, "game.listenPacket[onPacketCb]"); cbFn == nil {
					throwException("game.subscribePacket", errStr)
				} else {
					deRegFn, err := hb.RegPacketCallBack(str, func(pk packet.Packet) {
						strPk, err := json.Marshal(pk)
						if err != nil {
							printException("game.subscribePacket", "Cannot convert host packet to Json Str: "+err.Error())
							cbFn.Call(getReceiver(info), v8go.Null(iso))
						} else {
							val, err := v8go.JSONParse(info.Context(), string(strPk))
							if err != nil {
								printException("game.subscribePacket", "Cannot Parse Json Packet in Host: "+str)
								cbFn.Call(getReceiver(info), v8go.Null(iso))
								return
							} else {
								cbFn.Call(getReceiver(info), val)
							}
						}
					}, t)
					if err != nil {
						return throwException("game.subscribePacket", err.Error())
					}
					jsCbFn := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
						deRegFn()
						return nil
					})
					return jsCbFn.GetFunction(info.Context()).Value
				}
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function game.listenChat(onMsg function(name,msg)) deRegFn
	if err := game.Set("listenChat",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if errStr, cbFn := hasFuncIn(info, 0, "game.listenChat[onMsg]"); cbFn == nil {
				throwException("game.listenChat", errStr)
			} else {
				ctx := info.Context()
				deRegFn, err := hb.RegPacketCallBack("IDText", func(pk packet.Packet) {
					p := pk.(*packet.Text)
					SourceName, err := v8go.NewValue(iso, p.SourceName)
					if err != nil {
						printException("game.listenChat", err.Error())
						cbFn.Call(getReceiver(info), v8go.Null(iso), v8go.Null(iso))
						return
					}
					Message, err := v8go.NewValue(iso, p.Message)
					if err != nil {
						printException("game.listenChat", err.Error())
						cbFn.Call(getReceiver(info), v8go.Null(iso), v8go.Null(iso))
						return
					}
					cbFn.Call(getReceiver(info), SourceName, Message)
				}, t)
				if err != nil {
					return throwException("game.listenChat", err.Error())
				}
				jsCbFn := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
					deRegFn()
					return nil
				})
				t.TerminateHook = append(t.TerminateHook, deRegFn)
				return jsCbFn.GetFunction(ctx).Value
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	consts := v8go.NewObjectTemplate(iso)
	s256v, _ := v8go.NewValue(iso, identifyStr)
	consts.Set("script_sha256", s256v)
	consts.Set("script_path", scriptPath)
	consts.Set("engine_version", JSVERSION)
	//consts.Set("user_name","Not implemented")
	for k, v := range hb.GetQueries() {
		val, _ := v8go.NewValue(iso, v())
		consts.Set(k, val)
	}
	global.Set("consts", consts)
	/*
		// function FB_Query(info string) string
		if err:=global.Set("FB_Query",
			v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
				if str,ok:= hasStrIn(info,0,"FB_Query[info]"); !ok{
					throwException("FB_Query",str)
				}else{
					if str=="script_sha"{
						value,_:=v8go.NewValue(iso,identifyStr)
						return value
					}else if str=="script_path"{
						value,_:=v8go.NewValue(iso,scriptPath)
						return value
					}
					userInput:= hb.Query(str)
					value,_:=v8go.NewValue(iso,userInput)
					return value
				}
				return nil
			}),
		); err!=nil{panic(err)}*/

	fs := v8go.NewObjectTemplate(iso)
	global.Set("fs", fs)

	if err := fs.Set("getAbsPath",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if str, ok := hasStrIn(info, 0, "fs.getAbsPath[path]"); !ok {
				throwException("fs.getAbsPath", str)
			} else {
				absPath := hb.GetAbsPath(str)
				value, _ := v8go.NewValue(iso, absPath)
				return value
			}
			return nil
		})); err != nil {
		panic(err)
	}

	if err := fs.Set("requestFilePermission",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if hint, ok := hasStrIn(info, 1, "fs.requestFilePermission[hint]"); !ok {
				throwException("fs.requestFilePermission", hint)
				return nil
			} else {
				if dir, ok := hasStrIn(info, 0, "fs.requestFilePermission[hint]"); !ok {
					throwException("fs.requestFilePermission", dir)
					return nil
				} else {
					dir = hb.GetAbsPath(dir) + string(os.PathSeparator)
					if !AllowPath(dir) {
						throwException("fs.requestFilePermission", "The script is breaking out sandbox, aborting.")
						t.Terminate()
						return nil
					}
					permissionKey := "VisitDir:" + dir
					if hasPermission, ok := permission[permissionKey]; ok && hasPermission {
						value, _ := v8go.NewValue(iso, true)
						return value
					} else {
						for {
							warning := "Script[" + scriptName + "][" + _scriptName + "]wants to access the contents of directory " + dir + ".\n" +
								"Reason " + hint + "\n" +
								"(Warning: The script will gain the ability of REMOVING, MODIFYING, CREATING any file in this directory.)\n" +
								"Allow the access? Give an answer[y/N]:"
							choose := hb.GetInput(warning, t, scriptName)
							if choose == "Y" || choose == "y" {
								value, _ := v8go.NewValue(iso, true)
								permission[permissionKey] = true
								updatePermission()
								return value
							} else {
								value, _ := v8go.NewValue(iso, false)
								return value
							}
							//hb.Println("无效输入，请输入[是/否/Y/y/N/n]其中之一",t,scriptName)
						}
					}
				}
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function fs.readFile(path string) string
	// if permission is not granted or read fail, "" is returned
	if err := fs.Set("readFile",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if str, ok := hasStrIn(info, 0, "fs.readFile[path]"); !ok {
				throwException("fs.readFile", str)
			} else {
				p := hb.GetAbsPath(str)
				hasPermission := false
				for permissionName, _ := range permission {
					if strings.HasPrefix(permissionName, "VisitDir:") {
						if strings.HasPrefix(p, permissionName[len("VisitDir:"):]) {
							hasPermission = true
							break
						}
					}
				}
				if !hasPermission {
					throwException("fs.readFile", "The script is trying to access an external path (without permission), aborting.")
					t.Terminate()
					return nil
				}
				if !AllowPath(p) {
					throwException("fs.readFile", "The script is trying to access an external path (without permission), aborting.")
					t.Terminate()
					return nil
				}
				data, err := hb.LoadFile(p)
				if err != nil {
					value, _ := v8go.NewValue(iso, "")
					return value
				}
				value, _ := v8go.NewValue(iso, data)
				return value
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	// function fs.writeFile(path string,data string) isSuccess
	if err := fs.Set("writeFile",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if p, ok := hasStrIn(info, 0, "fs.writeFile[path]"); !ok {
				throwException("fs.writeFile", p)
			} else {
				if data, ok := hasStrIn(info, 1, "fs.writeFile[data]"); !ok {
					throwException("fs.writeFile", data)
				} else {
					p := hb.GetAbsPath(p)
					hasPermission := false
					for permissionName, _ := range permission {
						if strings.HasPrefix(permissionName, "VisitDir:") {
							if strings.HasPrefix(p, permissionName[len("VisitDir:"):]) {
								hasPermission = true
								break
							}
						}
					}
					if !hasPermission {
						throwException("fs.writeFile", "The script is trying to access an external path (without permission), aborting.")
						t.Terminate()
						return nil
					}
					if !AllowPath(p) {
						throwException("fs.writeFile", "The script is trying to access an external path (without permission), aborting.")
						t.Terminate()
						return nil
					}
					err := hb.SaveFile(p, data)
					if err != nil {
						value, _ := v8go.NewValue(iso, false)
						return value
					}
					value, _ := v8go.NewValue(iso, true)
					return value
				}
			}
			return nil
		})); err != nil {
		panic(err)
	}

	ws := v8go.NewObjectTemplate(iso)
	global.Set("ws", ws)

	if err := ws.Set("connect",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if address, ok := hasStrIn(info, 0, "ws.connect[address]"); !ok {
				throwException("ws.connect", address)
			} else {
				if errStr, cbFn := hasFuncIn(info, 1, "ws.connect[onNewMessage]"); cbFn == nil {
					throwException("ws.connect", errStr)
				} else {
					ctx := info.Context()
					conn, _, err := websocket.DefaultDialer.Dial(address, nil)
					if err != nil {
						return throwException("ws.connect", err.Error())
					}
					jsWriteFn := v8go.NewFunctionTemplate(iso, func(writeInfo *v8go.FunctionCallbackInfo) *v8go.Value {
						if t.Terminated() {
							return nil
						}
						if len(writeInfo.Args()) < 2 {
							throwException("SendMsg returned by FB_websocketConnectV2", "not enough arguments")
							return nil
						}
						if !writeInfo.Args()[1].IsString() {
							throwException("SendMsg returned by FB_websocketConnectV2", "SendMsg[data] should be string")
						}
						msgType := int(writeInfo.Args()[0].Number())
						err := conn.WriteMessage(msgType, []byte(writeInfo.Args()[1].String()))
						if err != nil {
							return throwException("SendMsg returned by FB_websocketConnectV2", "write fail")
						}
						return nil
					})
					go func() {
						for {
							msgType, data, err := conn.ReadMessage()
							if t.Terminated() {
								return
							}
							if err != nil {
								cbFn.Call(getReceiver(info), v8go.Null(iso), v8go.Null(iso))
								return
							}
							jsMsgType, err := v8go.NewValue(iso, int32(msgType))
							jsMsgData, err := v8go.NewValue(iso, string(data))
							cbFn.Call(getReceiver(info), jsMsgType, jsMsgData)
						}
					}()
					return jsWriteFn.GetFunction(ctx).Value
				}
			}
			return nil
		}),
	); err != nil {
		panic(err)
	}

	wsserve := func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		var address, pattern, errStr string
		var jsOnConnect *v8go.Function
		var ok bool
		if address, ok = hasStrIn(info, 0, "ws.serve[address]"); !ok {
			throwException("ws.serve", address)
			return nil
		}
		if pattern, ok = hasStrIn(info, 1, "ws.serve[pattern]"); !ok {
			throwException("ws.serve", pattern)
			return nil
		}
		if errStr, jsOnConnect = hasFuncIn(info, 2, "ws.serve[onConnect]"); jsOnConnect == nil {
			throwException("ws.serve", errStr)
			return nil
		}
		http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			wsConn, err := (&websocket.Upgrader{
				ReadBufferSize:  1024 * 10,
				WriteBufferSize: 1024 * 10,
			}).Upgrade(w, r, nil)
			if err != nil {
				throwException("ws.serve", err.Error())
				return
			}

			jsSendFn := v8go.NewFunctionTemplate(iso, func(writeInfo *v8go.FunctionCallbackInfo) *v8go.Value {
				if t.Terminated() {
					return nil
				}
				if len(writeInfo.Args()) < 2 {
					throwException("ws.serve.onConnect.sendMsg", "not enough arguments")
					return nil
				}
				if !writeInfo.Args()[1].IsString() {
					throwException("ws.serve.onConnect.sendMsg", "SendMsg[data] should be string")
				}
				msgType := int(writeInfo.Args()[0].Number())
				err := wsConn.WriteMessage(msgType, []byte(writeInfo.Args()[1].String()))
				if err != nil {
					return throwException("ws.serve.onConnect.sendMsg", "write fail")
				}
				return nil
			}).GetFunction(info.Context())
			jsCloseFn := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
				wsConn.Close()
				return nil
			}).GetFunction(info.Context())

			onMsgFn, err := jsOnConnect.Call(getReceiver(info), jsSendFn, jsCloseFn)
			if err != nil {
				throwException("ws.serve.onConnect.onMsgFn", err.Error())
				t.Terminate()
				return
			}
			jsOnMsgFn, err := onMsgFn.AsFunction()
			if err != nil {
				throwException("ws.serve.onConnect.onMsgFn", err.Error())
				t.Terminate()
				return
			}
			go func() {
				for {
					msgType, data, err := wsConn.ReadMessage()
					//fmt.Println("line 739")
					if err != nil {
						jsOnMsgFn.Call(getReceiver(info), v8go.Null(iso), v8go.Null(iso))
					} else {
						jsMsgType, _ := v8go.NewValue(iso, int32(msgType))
						jsMsgData, _ := v8go.NewValue(iso, string(data))
						jsOnMsgFn.Call(getReceiver(info), jsMsgType, jsMsgData)
					}
					//if err!=nil{
					//	println(err)
					//	return
					//}
					//fmt.Printf("%v %v\n",msgType,data)
					//err = wsConn.WriteMessage(msgType, data)
					//if err!=nil{
					//	println(err)
					//	return
					//}
				}
			}()
		})
		go http.ListenAndServe(address, nil)
		return nil
	}

	// ws.serve(address string,onNewMessage func(msgType int,data string)) func SendMsg(msgType int, data string)
	// 一般情况下，MessageType 为1(Text Messsage),即字符串类型，或者 0 byteArray (也被以字符串的方式传递)
	// onNewMessage 在连接关闭时会读取到两个null值
	if err := ws.Set("serve", v8go.NewFunctionTemplate(iso, wsserve)); err != nil {
		panic(err)
	}

	// fetch
	if err := fetch.InjectTo(iso, global); err != nil {
		panic(err)
	}
	// setTimeout, clearTimeout, setInterval and clearInterval
	if err := timers.InjectTo(iso, global); err != nil {
		panic(err)
	}
	//  atob and btoa
	if err := base64.InjectTo(iso, global); err != nil {
		panic(err)
	}

	// now we use built-in js, see built_in folder
	// encryption encryption.aesEncrypt(text, key)
	//encryption:=v8go.NewObjectTemplate(iso)
	//global.Set("encryption", encryption)
	//if err := encryption.Set("aesEncrypt",
	//	v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
	//		if text, ok := hasStrIn(info, 0, "encryption.aesEncrypt[text]"); !ok {
	//			throwException("encryption.aesEncrypt", text)
	//		} else {
	//			if key, ok := hasStrIn(info, 1, "encryption.aesEncrypt[key]"); !ok {
	//				throwException("encryption.aesEncrypt", key)
	//			} else {
	//				encryptOut,iv,err := aesEncrypt(text,key)
	//				if err!=nil{
	//					throwException("encryption.aesEncrypt",err.Error())
	//					return nil
	//				}else{
	//					result:=v8go.NewObjectTemplate(iso)
	//					jsEncryptOut, _ := v8go.NewValue(iso, encryptOut)
	//					jsIV, _ := v8go.NewValue(iso, iv)
	//					result.Set("cipherText",jsEncryptOut)
	//					result.Set("iv",jsIV)
	//					obj,_:=result.NewInstance(info.Context())
	//					return obj.Value
	//				}
	//			}
	//		}
	//		return nil
	//	}),
	//); err != nil {
	//	panic(err)
	//}
	//if err := encryption.Set("aesDecrypt",
	//	v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
	//		if text, ok := hasStrIn(info, 0, "encryption.aesDecrypt[text]"); !ok {
	//			throwException("encryption.aesDecrypt", text)
	//		} else {
	//			if key, ok := hasStrIn(info, 1, "encryption.aesDecrypt[key]"); !ok {
	//				throwException("encryption.aesDecrypt", key)
	//			} else {
	//				if iv, ok := hasStrIn(info, 2, "encryption.aesDecrypt[iv]"); !ok {
	//					throwException("encryption.aesDecrypt", key)
	//				} else{
	//					decryptOut,err := aesDecrypt(text,key,iv)
	//					if err!=nil{
	//						throwException("encryption.aesDecrypt",err.Error())
	//						return nil
	//					}else{
	//						value, _ := v8go.NewValue(iso, decryptOut)
	//						return value
	//					}
	//				}
	//			}
	//		}
	//		return nil
	//	}),
	//); err != nil {
	//	panic(err)
	//}

	return func() {
		t.Terminate()
	}
}

func CtxFunctionInject(ctx *v8go.Context) {
	// URL and URLSearchParams
	if err := url.InjectTo(ctx); err != nil {
		panic(err)
	}
	_, err := ctx.RunScript(built_in.GetbuiltIn(), "built_in")
	if err != nil {
		panic(err)
	}
}
