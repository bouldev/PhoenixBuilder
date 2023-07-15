package system

import (
	"phoenixbuilder/solutions/omega_lua/omega_lua/pollers"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type SystemModule struct {
	LuaGoSystem
	pollers.LuaAsyncInvoker
}

type LuaGoSystem interface {
	Print(string)
	UserInputChan() <-chan string
}

func NewSystemModule(goImplements LuaGoSystem, luaAsyncInvoker pollers.LuaAsyncInvoker) *SystemModule {
	return &SystemModule{
		LuaGoSystem:     goImplements,
		LuaAsyncInvoker: luaAsyncInvoker,
	}
}

func (m *SystemModule) MakeLValue(L *lua.LState) (lua.LValue, map[lua.LValue]pollers.LuaEventDataChanMaker) {
	luaModule := L.NewTable()
	startTime := float64(time.Now().UnixMilli()) / 1000
	luaModule = L.SetFuncs(luaModule, map[string]lua.LGFunction{
		"print": m.luaGoSystemPrint,
		"os":    m.luaGoSystemOs,
		"cwd":   m.luaGoSystemCwd,
		"now": func(l *lua.LState) int {
			l.Push(lua.LNumber((float64(time.Now().UnixMilli()) / 1000) - startTime))
			return 1
		},
	})
	// poller flags for sleep and input
	flagSleep := L.NewFunction(m.luaGoSleep)
	flagInput := L.NewFunction(m.luaGoInput)
	pollerFlags := map[lua.LValue]pollers.LuaEventDataChanMaker{
		flagSleep: goSleepSourceMaker,
		flagInput: m.goInputSourceMaker,
	}
	// inject block_sleep and block_input flags into module
	luaModule.RawSetString("block_sleep", flagSleep)
	luaModule.RawSetString("block_input", flagInput)
	// inject start_time into module
	luaModule.RawSetString("start_time", lua.LNumber(startTime))
	return luaModule, pollerFlags
}
