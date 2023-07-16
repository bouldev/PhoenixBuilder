package omega_lua

import (
	"context"
	lua "github.com/yuin/gopher-lua"
	"phoenixbuilder/solutions/omega_lua/omega_lua/concurrent"
	"phoenixbuilder/solutions/omega_lua/omega_lua/lua_utils"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/command"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/listen"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/packets_utils"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/system"
	submodule_holder "phoenixbuilder/solutions/omega_lua/omega_lua/modules_holder"
)

type GoImplements struct {
	GoSystem         system.LuaGoSystem
	GoPackets        packets_utils.LuaGoPackets
	GoPacketProvider listen.PacketProvider
	GoCmdSender      command.CmdSender
}

func CreateOmegaLuaEnv(ctx context.Context,
	goImplements *GoImplements,
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

	// pollers
	ListenModule := listen.NewListenModule(ac,
		goImplements.GoPacketProvider, packetsModule.NewGamePacket,
		systemPollerFlags)
	luaListenModule := ListenModule.MakeLValue(L)

	// load modules
	L.PreloadModule("omega", submodule_holder.NewSubModuleHolder(map[string]lua.LValue{
		"system":  luaSystemModule,
		"listen":  luaListenModule,
		"packets": luaPacketsModule,
		"cmds":    luaCmdModule,
	}).Loader)
	return ac, L
}
