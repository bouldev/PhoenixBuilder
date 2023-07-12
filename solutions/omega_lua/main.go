package main

import (
	"context"
	"phoenixbuilder/fastbuilder/lib/utils/file_wrapper"
	"phoenixbuilder/solutions/omega_lua/monk"
	"phoenixbuilder/solutions/omega_lua/omega_lua/concurrent"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/backend"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/listen"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/packets_utils"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/system"
	submodule_holder "phoenixbuilder/solutions/omega_lua/omega_lua/modules_holder"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// 根据指定的目录读取消息 返回值为读取后的字符串 与 一个err
func ReadOutLuaCodeFromFile(fileName string) (code string, err error) {
	data, err := file_wrapper.GetFileData(fileName)
	if err != nil {
		return "", err
	}
	code = string(data)
	return
}

func main() {
	// read lua
	//测试用读取的packet.lua
	code, err := ReadOutLuaCodeFromFile("packet.lua")
	if err != nil {
		panic(err)
	}
	if code == "" {
		panic("empty lua code")
	}
	// create lua state
	L := lua.NewState()
	defer L.Close()
	// create async ctrl
	//创建一个async控制对象
	ac := concurrent.NewAsyncCtrl(context.Background())
	//创建一个packetSize为128的游戏监听器
	monkListener := monk.NewMonkListen(128)
	//向omega中注册内置table
	//makeLvalue即是注册这个子表中的各种属性
	luaPacketsModule := packets_utils.NewOmegaPacketsModule(monkListener)
	L.PreloadModule("omega", submodule_holder.NewSubModuleHolder(map[string]lua.LValue{
		"backend": backend.NewOmegaBackendModule(&monk.MonkBackend{}).MakeLValue(L),
		"system":  system.NewOmegaSystemModule(ac).MakeLValue(L),
		"packets": luaPacketsModule.MakeLValue(L),
		"block": listen.NewOmegaBlockModule(
			ac,
			monk.NewMonkListen(128),
			luaPacketsModule,
		).MakeLValue(L),
	}).Loader)
	// run lua code
	errChan := concurrent.FireLuaCodeInGoRoutine(ac, L, code)
	// wait for lua code to finish
	time.Sleep(time.Second * 2)
	err = <-errChan
	if err != nil {
		panic(err)
	}
}
