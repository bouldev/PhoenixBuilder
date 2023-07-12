package listen

import (
	"context"
	"phoenixbuilder/solutions/omega_lua/omega_lua/concurrent"
	"phoenixbuilder/solutions/omega_lua/omega_lua/modules/packets_utils"

	lua "github.com/yuin/gopher-lua"
)

// block module
type OmegaBlockModule struct {
	goImplements    LuaGoListen
	luaFns          map[lua.LValue]string
	ac              *concurrent.AsyncCtrl
	luaPacketModule *packets_utils.OmegaPacketsModule
}

func NewOmegaBlockModule(
	ac *concurrent.AsyncCtrl,
	goImplements LuaGoListen,
	luaPacketModule *packets_utils.OmegaPacketsModule,
) *OmegaBlockModule {
	m := &OmegaBlockModule{
		ac:              ac,
		goImplements:    goImplements,
		luaPacketModule: luaPacketModule,
	}
	return m
}

func (m *OmegaBlockModule) MakeLValue(L *lua.LState) lua.LValue {
	blockModule := L.NewTable()
	m.luaFns = map[lua.LValue]string{
		//实现睡眠函数
		L.NewFunction(m.luaGoBlockSleep): LuaListenFnNameSleep,
		//实现获取用户输入
		L.NewFunction(m.luaGoBlockGetUserInput): LuaListenFnNameGetUserInput,
	}
	for fn, name := range m.luaFns {
		L.SetField(blockModule, name, fn)
	}
	L.SetField(blockModule, "make_mux_poller", L.NewFunction(m.luaGoBlockPoller))
	L.SetField(blockModule, "make_packet_poller", L.NewFunction(m.luaGoNewPacketPoller))
	registerBlockPoller(L)
	registerPacketPoller(L)
	return blockModule
}

// block.make_mux_poller()
func (m *OmegaBlockModule) luaGoBlockPoller(L *lua.LState) int {
	goPoller := newBlockPoller(m, L, m.ac.Context())
	luaPoller := goPoller.MakeLValue(L)
	L.Push(luaPoller)
	goPoller.luaSelf = luaPoller
	return 1
}

// block module: make_packet_poller(type1,type2,...)
func (m *OmegaBlockModule) luaGoNewPacketPoller(L *lua.LState) int {
	packetTypes := make([]string, 0, L.GetTop())
	for i := 1; i <= L.GetTop(); i++ {
		packetTypes = append(packetTypes, L.ToString(i))
	}
	ctx, cancelFn := context.WithCancel(m.ac.Context())
	feedChan := m.goImplements.MakeMCPacketFeeder(ctx, packetTypes)
	goPoller := newPacketPoller(m, feedChan, ctx, cancelFn)
	luaPoller := goPoller.MakeLValue(L)
	L.Push(luaPoller)
	return 1
}
