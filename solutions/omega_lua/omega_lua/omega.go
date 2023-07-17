package omega_lua

import (
	"context"
	"phoenixbuilder/solutions/omega_lua/omega_lua/concurrent"
	"phoenixbuilder/solutions/omega_lua/omega_lua/lua_utils"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/command"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/listen"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/packets_utils"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/system"
	submodule_holder "phoenixbuilder/solutions/omega_lua/omega_lua/modules_holder"

	lua "github.com/yuin/gopher-lua"
)

type GoImplements struct {
	GoSystem         system.LuaGoSystem
	GoPackets        packets_utils.LuaGoPackets
	GoPacketProvider listen.PacketProvider
	GoCmdSender      command.CmdSender
}

func CreateOmegaLuaEnv(ctx context.Context,
	goImplements *GoImplements,
	config *lua_utils.LuaConfigRaw,
	baseStoragePath string,
) (ac concurrent.AsyncCtrl, L *lua.LState) {
	L = lua.NewState()
	ac = concurrent.NewAsyncCtrl(ctx)
	// go implements
	// 1. monk system
	// lua wrapper
	lua_utils.RegisterCommonLuaGoObject(L)
	systemModule := system.NewSystemModule(goImplements.GoSystem, ac)
	luaSystemModule, systemPollerFlags := systemModule.MakeLValue(L)
	packetsModule := packets_utils.NewOmegaPacketsModule(goImplements.GoPackets)
	luaPacketsModule := packetsModule.MakeLValue(L)
	cmdModule := command.NewCmdModule(goImplements.GoCmdSender, packetsModule.NewGamePacket)
	luaCmdModule := cmdModule.MakeLValue(L, ac)
	storageModule := lua_utils.NewStorageModule(baseStoragePath)
	luaStorageModule := storageModule.MakeLValue(L)

	// pollers
	ListenModule := listen.NewListenModule(ac,
		goImplements.GoPacketProvider, packetsModule.NewGamePacket,
		systemPollerFlags)
	luaListenModule := ListenModule.MakeLValue(L)

	// config
	goConfig := config.Config
	luaConfigUpgradeFn := func(L *lua.LState) int {
		d := lua_utils.CommonGoObjectFromUserData(L)
		luaConfig := lua_utils.CheckCommonGoObject(L)
		newConfig := luaConfig.GetData()
		newConfig = lua_utils.CleanMapKey(newConfig)
		config.OnConfigUpdate(newConfig)
		return d
	}
	luaConfig := lua_utils.NewCommonLuaGoObject(goConfig, map[string]lua.LValue{
		"upgrade": L.NewFunction(luaConfigUpgradeFn),
	}).MakeLValue(L)

	// load modules
	L.PreloadModule("omega", submodule_holder.NewSubModuleHolder(map[string]lua.LValue{
		"system":  luaSystemModule,
		"listen":  luaListenModule,
		"packets": luaPacketsModule,
		"cmds":    luaCmdModule,
		"config":  luaConfig,
		"storage": luaStorageModule,
	}).Loader)
	L.SetGlobal("ud2lua", L.NewFunction(lua_utils.UserDataToLuaValue))
	return ac, L
}
