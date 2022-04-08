package script_engine

import (
	"context"
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
	
	"regexp"
	"io/ioutil"
	"path/filepath"
	"errors"
)

// jsEngine.hostBridge.api
const JSVERSION = "[script_engine@v8].gamma.7"

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
	return info.Context().Global()
}

func InitHostFns(iso *v8go.Isolate, global *v8go.ObjectTemplate, hb bridge.HostBridge, _scriptName string, identifyStr string, scriptPath string, bundle *ScriptPackage) func() {
	scriptName := _scriptName
	//permission := LoadPermission(hb, identifyStr)
	/*updatePermission := func() {
		SavePermission(hb, identifyStr, permission)
	}*/

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
	global.Set("printf",v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args:=info.Args()
		if(len(args)==0) {
			// Same to printf("");
			fmt.Printf("[%s] ",scriptName)
			return nil
		}
		things:=make([]interface{},len(args)-1)
		for i, v:=range args[1:] {
			things[i]=v.String()
		}
		fmt.Printf(fmt.Sprintf("[%s] %s",scriptName,args[0].String()),things...)
		return nil
	}))
	global.Set("sprintf",v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args:=info.Args()
		if(len(args)==0) {
			// Same to sprintf("");
			val,_ := v8go.NewValue(iso, "")
			return val
		}
		things:=make([]interface{},len(args)-1)
		for i, v:=range args[1:] {
			things[i]=v.String()
		}
		val, _:=v8go.NewValue(iso, fmt.Sprintf(args[0].String(),things...))
		return val
	}))
	internalNameSet=false
	
	if err := engine.Set("setName",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if(internalNameSet) {
				throwException("engine.setName","Changing name dynamically for a package is not allowed")
				return nil
			}
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
	
	if err := engine.Set("setNameInternal",
		v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			if str, ok := hasStrIn(info, 0, "engine.setNameInternal[scriptName]"); !ok {
				throwException("engine.setNameInternal: No arguments assigned", str)
			} else {
				internalNameSet=true
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
	// Implemented in built-in js

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
				if errStr, cbFn := hasFuncIn(info, 1, "game.subscribePacket[onPacketCb]"); cbFn == nil {
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
	containerPath:=""

	consts := v8go.NewObjectTemplate(iso)
	//s256v, _ := v8go.NewValue(iso, identifyStr)
	//consts.Set("script_sha256", s256v)
	consts.Set("engine_version", JSVERSION)
	for k, v := range hb.GetQueries() {
		val, _ := v8go.NewValue(iso, v())
		consts.Set(k, val)
	}
	if(bundle!=nil) {
		bundleObj:=v8go.NewObjectTemplate(iso)
		bundleObj.Set("identifier", bundle.Identifier)
		bundleObj.Set("name", bundle.Name)
		bundleObj.Set("description", bundle.Description)
		bundleObj.Set("author", bundle.Author)
		bundleObj.Set("version", bundle.Version)
		bundleObj.Set("manifest", bundle.Manifest)
		bundleObj.Set("related_information", bundle.RelatedInformation)
		bundleObj.Set("entrypoint", bundle.EntryPoint)
		bundleContentObj:=v8go.NewObjectTemplate(iso)
		for n, v:=range bundle.Datas {
			bundleContentObj.Set(n,v)
		}
		for n, v:=range bundle.Scripts {
			csn:=string(n)
			csv:=v
			externalScriptObj:=v8go.NewObjectTemplate(iso)
			externalScriptObj.Set("run", v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
				newContext:=v8go.NewContext(iso, global)
				// Share the same `global`.
				nconsts,_:=newContext.Global().Get("consts")
				nconstso,_:=nconsts.AsObject()
				nbundle,_:=nconstso.Get("bundle")
				nbundleo,_:=nbundle.AsObject()
				nbundleo.Set("currentScript", csn)
				nbundleo.Set("fromRequire", true)
				nfs,_:=newContext.Global().Get("fs")
				nfso,_:=nfs.AsObject()
				nfso.Set("containerPath", containerPath)
				CtxFunctionInject(newContext)
				_, err:=csv.Run(newContext)
				if(err!=nil) {
					je:=err.(*v8go.JSError)
					throwException("runScript",fmt.Sprintf("Uncaught Error in script '%s': %s\nLocation: %s\nStack: %s",n,je.Message,je.Location,je.StackTrace))
					return nil
				}
				n_mdl,err:=newContext.Global().Get("module")
				if(err!=nil) {
					return nil
				}
				n_mdl_obj,err:=n_mdl.AsObject()
				if(err!=nil) {
					return nil
				}
				n_exp,err:=n_mdl_obj.Get("exports")
				if(err!=nil) {
					return nil
				}
				return n_exp
			}))
			bundleContentObj.Set(n,externalScriptObj)
		}
		bundleObj.Set("content",bundleContentObj)
		consts.Set("bundle", bundleObj)
	}
	global.Set("consts", consts)
	
	module:=v8go.NewObjectTemplate(iso)
	module.Set("exports", v8go.NewObjectTemplate(iso))
	global.Set("module", module)

	fs := v8go.NewObjectTemplate(iso)
	global.Set("fs", fs)
	
	fs.Set("containerPath", "")

	fs.Set("requireContainer", v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if(len(containerPath)!=0) {
			throwException("fs.requireContainer", "Requesting a container, but a container for the script is already created.")
			return nil
		}
		if(len(info.Args())==0) {
			throwException("fs.requireContainer", "No arguments provided")
			return nil
		}
		containerNameValue:=info.Args()[0]
		if(!containerNameValue.IsString()) {
			throwException("fs.requireContainer", "Container name should be a string, for example, com.user.myscript")
			return nil
		}
		containerName:=containerNameValue.String()
		if(len(containerName)>32||len(containerName)<4) {
			throwException("fs.requireContainer", "Container name is too long or too short")
			return nil
		}
		containerExpection:=regexp.MustCompile("^([A-Za-z0-9_-]|\\.)*$")
		if(!containerExpection.MatchString(containerName)) {
			throwException("fs.requireContainer", "Invalid container name! Container name should be in this format: com.user.myscript")
			return nil
		}
		homedir, _:=os.UserHomeDir()
		containerPath=filepath.Join(homedir, fmt.Sprintf(".config/fastbuilder/containers/%s", containerName)) + "/"
		os.MkdirAll(containerPath, 0700)
		containerpv,_:=v8go.NewValue(iso, containerPath)
		info.This().Set("containerPath",containerPath)
		return containerpv
	}))
	
	fs.Set("exists", v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if(len(containerPath)==0) {
			throwException("fs.exists", "File operation without a container created")
			return nil
		}
		if(len(info.Args())==0) {
			throwException("fs.exists", "No arguments provided")
			return nil
		}
		pathv:=info.Args()[0]
		if(!pathv.IsString()) {
			throwException("fs.exists", "Argument \"path\" must be a string.")
			return nil
		}
		path:=filepath.Clean(pathv.String())
		if(path[0]!='/') {
			path=fmt.Sprintf("%s%s",containerPath,path)
		}else{
			if(path[0:len(containerPath)]!=containerPath) {
				throwException("fs.exists", "Trying to control filesystem out of container")
>>>>>>> f168b46 (Sandbox & script bundle)
				return nil
			}
		}
		_, err:=os.Stat(path)
		r:=false
		if(!errors.Is(err, os.ErrNotExist)) {
			r=true
		}
		rv,_:=v8go.NewValue(iso,r)
		return rv
	}))
	
	fs.Set("isDir", v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if(len(containerPath)==0) {
			throwException("fs.isDir", "File operation without a container created")
			return nil
		}
		if(len(info.Args())==0) {
			throwException("fs.isDir", "No arguments provided")
			return nil
		}
		pathv:=info.Args()[0]
		if(!pathv.IsString()) {
			throwException("fs.isDir", "Argument \"path\" must be a string.")
			return nil
		}
		path:=filepath.Clean(pathv.String())
		if(path[0]!='/') {
			path=fmt.Sprintf("%s%s",containerPath,path)
		}else{
			if(path[0:len(containerPath)]!=containerPath) {
				throwException("fs.isDir", "Trying to control filesystem out of container")
				return nil
			}
		}
		st, err:=os.Stat(path)
		r:=false
		if(err==nil) {
			if(st.IsDir()) {
				r=true
			}
		}
		rv,_:=v8go.NewValue(iso,r)
		return rv
	}))
	
	fs.Set("mkdir", v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if(containerPath=="") {
			throwException("fs.mkdir", "Container not created")
			return nil
		}
		if(len(info.Args())==0) {
			throwException("fs.mkdir", "No arguments provided")
			return nil
		}
		pathv:=info.Args()[0]
		if(!pathv.IsString()) {
			throwException("fs.mkdir", "Argument \"path\" must be a string.")
			return nil
		}
		path:=filepath.Clean(pathv.String())
		if(path[0]!='/') {
			path=fmt.Sprintf("%s%s",containerPath,path)
		}else{
			if(path[0:len(containerPath)]!=containerPath) {
				throwException("fs.mkdir", "Trying to control filesystem out of container")
				return nil
			}
		}
		err:=os.MkdirAll(containerPath, 0700)
		if(err!=nil) {
			throwException("fs.mkdir", fmt.Sprintf("Failed to perform operation: %v",err))
		}
		return nil
	}))
	
	fs.Set("rename", v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if(containerPath=="") {
			throwException("fs.rename", "Container not created")
			return nil
		}
		if(len(info.Args())<2) {
			throwException("fs.rename", "Required 2 arguments")
			return nil
		}
		pathv:=info.Args()[0]
		npathv:=info.Args()[1]
		if(!pathv.IsString()) {
			throwException("fs.rename", "Argument \"oldpath\" must be a string.")
			return nil
		}
		if(!npathv.IsString()) {
			throwException("fs.rename", "Argument \"newpath\" must be a string.")
			return nil
		}
		path:=filepath.Clean(pathv.String())
		npath:=filepath.Clean(npathv.String())
		if(path[0]!='/') {
			path=fmt.Sprintf("%s%s",containerPath,path)
		}else{
			if(path[0:len(containerPath)]!=containerPath) {
				throwException("fs.rename", "Trying to control filesystem out of container")
				return nil
			}
			if(path==containerPath) {
				throwException("fs.rename", "Controlling container is not allowed")
				return nil
			}
		}
		if(npath[0]!='/') {
			npath=fmt.Sprintf("%s%s",containerPath,npath)
		}else{
			if(npath[0:len(containerPath)]!=containerPath) {
				throwException("fs.rename", "Trying to control filesystem out of container")
				return nil
			}
			if(npath==containerPath) {
				throwException("fs.rename", "Controlling container is not allowed")
				return nil
			}
		}
		err:=os.Rename(path,npath)
		if(err!=nil) {
			throwException("fs.rename", fmt.Sprintf("Failed to perform operation: %v",err))
		}
		return nil
	}))
	
	fs.Set("remove", v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if(containerPath=="") {
			throwException("fs.remove", "Container not created")
			return nil
		}
		if(len(info.Args())==0) {
			throwException("fs.remove", "No arguments provided")
			return nil
		}
		pathv:=info.Args()[0]
		if(!pathv.IsString()) {
			throwException("fs.remove", "Argument \"path\" must be a string.")
			return nil
		}
		path:=filepath.Clean(pathv.String())
		if(path[0]!='/') {
			path=fmt.Sprintf("%s%s",containerPath,path)
		}else{
			if(path[0:len(containerPath)]!=containerPath) {
				throwException("fs.remove", "Trying to control filesystem out of container")
				return nil
			}
			if(path==containerPath) {
				throwException("fs.remove", "Removing container is not allowed")
				return nil
			}
		}
		err:=os.RemoveAll(containerPath)
		if(err!=nil) {
			throwException("fs.remove", fmt.Sprintf("Failed to perform operation: %v",err))
		}
		return nil
	}))
	
	fs.Set("readFile", v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if(len(containerPath)==0) {
			throwException("fs.readFile", "File operation without a container created")
			return nil
		}
		if(len(info.Args())==0) {
			throwException("fs.readFile", "No arguments provided")
			return nil
		}
		pathv:=info.Args()[0]
		if(!pathv.IsString()) {
			throwException("fs.readFile", "Argument \"path\" must be a string.")
			return nil
		}
		path:=filepath.Clean(pathv.String())
		if(path[0]!='/') {
			path=fmt.Sprintf("%s%s",containerPath,path)
		}else{
			if(path[0:len(containerPath)]!=containerPath) {
				throwException("fs.mkdir", "Trying to control filesystem out of container")
				return nil
			}
		}
		filecontent, err := ioutil.ReadFile(path)
		if(err!=nil) {
			throwException("fs.readFile", fmt.Sprintf("Failed to read target file: %v", err))
			return nil
		}
		fin, _ := v8go.NewValue(iso, string(filecontent))
		return fin
	}))
	
	fs.Set("writeFile", v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		if(len(containerPath)==0) {
			throwException("fs.writeFile", "File operation without a container created")
		}
		if(len(info.Args())<2) {
			throwException("fs.writeFile", "Required 2 arguments")
			return nil
		}
		pathv:=info.Args()[0]
		content:=info.Args()[1]
		if(!pathv.IsString()) {
			throwException("fs.writeFile", "Argument \"path\" must be a string.")
			return nil
		}
		path:=filepath.Clean(pathv.String())
		if(path[0]!='/') {
			path=fmt.Sprintf("%s%s",containerPath,path)
		}else{
			if(path[0:len(containerPath)]!=containerPath) {
				throwException("fs.writeFile", "Trying to control filesystem out of container")
				return nil
			}
		}
		err:=ioutil.WriteFile(path,[]byte(content.String()), 0700)
		if(err!=nil) {
			throwException("fs.writeFile", fmt.Sprintf("Failed to write target file: %v", err))
			return nil
		}
		return nil
	}))
	
	newws := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		funcName := "new ws()"
		if !info.IsConstructCall {
			throwException(funcName, "Cannot call constructor as function")
			return nil
		}
		thisObject := info.This()
		if len(info.Args()) == 0 {
			throwException(funcName, "No arguments specified")
			return nil
		}
		address := info.Args()[0]
		if !address.IsString() {
			throwException(funcName, "argument address should be a String value.")
			return nil
		}
		thisObject.Set("isConnecting", true)
		go func() {
			ctx := info.Context()
			conn, _, err := websocket.DefaultDialer.Dial(address.String(), nil)
			if err != nil {
				onerror_obj, _ := thisObject.Get("onerror")
				if onerror_obj == nil {
					throwException(funcName, err.Error())
					return
				} else {
					onerror, _ := onerror_obj.AsFunction()
					if onerror == nil {
						throwException(funcName, err.Error())
						return
					}
					errorStr, _ := v8go.NewValue(iso, err.Error())
					onerror.Call(ctx.Global(), errorStr)
				}
			}
			closeFn := v8go.NewFunctionTemplate(iso, func(_ *v8go.FunctionCallbackInfo) *v8go.Value {
				conn.Close()
				thisObject.Set("closed", true)
				/*onclosefunc, _:=thisObject.Get("onclose")
				if onclosefunc == nil {
					return nil
				}
				onclosefn,_:=onclosefunc.AsFunction()
				if onclosefn==nil {
					return nil
				}
				onclosefn.Call(ctx.Global())*/
				return nil
			})
			jsWriteFn := v8go.NewFunctionTemplate(iso, func(writeInfo *v8go.FunctionCallbackInfo) *v8go.Value {
				if t.Terminated() {
					return nil
				}
				if len(writeInfo.Args()) < 1 {
					throwException("ws.send", "no enough arguments")
					return nil
				}
				if !writeInfo.Args()[0].IsString() {
					throwException("ws.send", "[data] should be string")
					return nil
				}
				msgType := 1
				if len(writeInfo.Args()) >= 2 {
					if !writeInfo.Args()[1].IsNumber() {
						throwException("ws.send", "non-number argument 1")
						return nil
					}
					msgType = int(writeInfo.Args()[1].Number())
				}
				err := conn.WriteMessage(msgType, []byte(writeInfo.Args()[0].String()))
				if err != nil {
					return throwException("ws.send", "Failed to write.")
				}
				return nil
			})
			thisObject.Set("send", jsWriteFn.GetFunction(ctx).Value)
			thisObject.Set("close", closeFn.GetFunction(ctx).Value)
			_onconn, _ := thisObject.Get("onconnection")
			if _onconn == nil {
				_onconn, _ = thisObject.Get("onopen")
			}
			thisObject.Set("isConnecting", false)
			if _onconn != nil {
				onconn, _ := _onconn.AsFunction()
				onconn.Call(ctx.Global(), info.This())
			}

			for {
				msgType, data, err := conn.ReadMessage()
				if t.Terminated() {
					return
				}
				__onmessage, _ := thisObject.Get("onmessage")
				__onclose, _ := thisObject.Get("onclose")
				var onmessage *v8go.Function
				var onclose *v8go.Function
				if __onmessage != nil {
					onmessage, _ = __onmessage.AsFunction()
				}
				if __onclose != nil {
					onclose, _ = __onclose.AsFunction()
				}
				if err != nil {
					thisObject.Set("closed", true)
					if onclose == nil {
						//throwException("ws Loop", fmt.Sprintf("Unhandled error, can be handled by setting ws.onerror: Error reading: %v",err))
						return
					}
					eStr, _ := v8go.NewValue(iso, fmt.Sprintf("%v", err))
					onclose.Call(info.Context().Global(), eStr)
					//cbFn.Call(getReceiver(info), v8go.Null(iso), v8go.Null(iso))
					return
				}
				if onmessage == nil {
					return
				}
				jsMsgType, err := v8go.NewValue(iso, int32(msgType))
				jsMsgData, err := v8go.NewValue(iso, string(data))
				onmessage.Call(info.Context().Global(), jsMsgData, jsMsgType)
			}
		}()
		return thisObject.Value
	})
	global.Set("ws", newws)
	wsServer := v8go.NewFunctionTemplate(iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		funcName := "new ws.Server(address)"
		if !info.IsConstructCall {
			throwException(funcName, "Cannot call constructor as function")
			return nil
		}
		args := info.Args()
		if len(args) == 0 {
			throwException(funcName, "Argument address is required!")
			return nil
		}
		if !args[0].IsString() {
			throwException(funcName, "Arguments should be of type String.")
			return nil
		}
		thisObject := info.This()
		ctx := info.Context()
		httpHandler := &mutableHandler{
			HTTPHandler: func(w http.ResponseWriter, r *http.Request) {
				wsConn, err := (&websocket.Upgrader{
					ReadBufferSize:  1024,
					WriteBufferSize: 1024,
				}).Upgrade(w, r, nil)
				if err != nil {
					// Connection haven't been established
					// So it's not an error that should be trapped.
					return
				}
				clientObject, _ := v8go.NewObjectTemplate(iso).NewInstance(ctx)
				clientObject.Set("path", r.URL.Path)
				onConn, _ := thisObject.Get("onconnection")
				if onConn == nil {
					// Maybe the script haven't get ready yet.
					wsConn.Close()
					return
				}
				onConnFn, _ := onConn.AsFunction()
				if onConnFn == nil {
					wsConn.Close()
					throwException("ws.Server/onConnection", ".onconnection isn't of type function.")
					return
				}
				sendMessageFnTemplate := v8go.NewFunctionTemplate(iso, func(info_ *v8go.FunctionCallbackInfo) *v8go.Value {
					iArgs := info_.Args()
					if len(iArgs) == 0 {
						throwException("ws.Server/client.send(msg,msgType)", "No arguments specified.")
						return nil
					}
					msgType := 1
					if len(iArgs) > 1 {
						if !iArgs[1].IsNumber() {
							throwException("ws.Server/client.send(msg,msgType)", "msgType: Number !")
							return nil
						}
						msgTypeVal := iArgs[1].Number()
						msgType = int(msgTypeVal)
					}
					if !iArgs[0].IsString() {
						throwException("ws.Server/client.send(msg,msgType)", "msg: String !")
						return nil
					}
					err := wsConn.WriteMessage(msgType, []byte(iArgs[0].String()))
					if err != nil {
						throwException("ws.Server/client.send(...)", fmt.Sprintf("Failed to write: %v", err))
					}
					return nil
				})
				terminateFT := v8go.NewFunctionTemplate(iso, func(_ *v8go.FunctionCallbackInfo) *v8go.Value {
					wsConn.Close()
					clientObject.Set("closed", true)
					return nil
				})
				clientObject.Set("send", sendMessageFnTemplate.GetFunction(ctx).Value)
				clientObject.Set("isConnecting", false)
				clientObject.Set("closed", false)
				clientObject.Set("close", terminateFT.GetFunction(ctx).Value)
				onConnFn.Call(ctx.Global(), clientObject)
				for {
					msgType, data, err := wsConn.ReadMessage()
					if t.Terminated() {
						return
					}
					__onmessage, _ := clientObject.Get("onmessage")
					__onclose, _ := clientObject.Get("onclose")
					var onmessage *v8go.Function
					var onclose *v8go.Function
					if __onmessage != nil {
						onmessage, _ = __onmessage.AsFunction()
					}
					if __onclose != nil {
						onclose, _ = __onclose.AsFunction()
					}
					if err != nil {
						clientObject.Set("closed", true)
						if onclose == nil {
							//throwException("ws Loop", fmt.Sprintf("Unhandled error, can be handled by setting ws.onerror: Error reading: %v",err))
							return
						}
						eStr, _ := v8go.NewValue(iso, fmt.Sprintf("%v", err))
						onclose.Call(ctx.Global(), eStr)
						//cbFn.Call(getReceiver(info), v8go.Null(iso), v8go.Null(iso))
						return
					}
					if onmessage == nil {
						return
					}
					jsMsgType, err := v8go.NewValue(iso, int32(msgType))
					jsMsgData, err := v8go.NewValue(iso, string(data))
					onmessage.Call(ctx.Global(), jsMsgData, jsMsgType)
				}
			},
		}
		server := http.Server{
			Addr:    args[0].String(),
			Handler: httpHandler,
		}
		shutdownServerFuncTemplate := v8go.NewFunctionTemplate(iso, func(_ *v8go.FunctionCallbackInfo) *v8go.Value {
			server.Shutdown(context.Background())
			return nil
		})
		thisObject.Set("shutdown", shutdownServerFuncTemplate)
		go func() {
			server.ListenAndServe()
			onServerShutdownObj, _ := thisObject.Get("onServerShutdown")
			if onServerShutdownObj == nil {
				return
			}
			onServerShutdown, _ := onServerShutdownObj.AsFunction()
			if onServerShutdown == nil {
				return
			}
			onServerShutdown.Call(info.Context().Global())
		}()
		return thisObject.Value
	})
	newws.Set("Server", wsServer, v8go.ReadOnly)

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
		e := err.(*v8go.JSError)
		fmt.Printf("Builtin Script ran into a runtime error, stack dump:\n")
		fmt.Printf("%s\n", e.Message)
		fmt.Printf("%s\n", e.Location)
		fmt.Printf("%s\n", e.StackTrace)
		panic(err)
	}
}
