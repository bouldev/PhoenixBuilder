package command

import (
	"phoenixbuilder/solutions/omega_lua/omega_lua/pollers"

	lua "github.com/yuin/gopher-lua"
)

type CmdRespHandler struct {
	*pollers.BasicDispatcher
	luaSelf lua.LValue
}

func NewCmdRespHandler(
	eventChan pollers.EventChan,
	luaAsyncInvoker pollers.LuaAsyncInvoker,
) *CmdRespHandler {
	p := &CmdRespHandler{
		BasicDispatcher: pollers.NewBasicDispatcher(eventChan, luaAsyncInvoker),
	}
	return p
}

func (m *CmdRespHandler) MakeLValue(L *lua.LState) lua.LValue {
	luaPoller := L.NewUserData()
	luaPoller.Value = m
	L.SetMetatable(luaPoller, L.GetTypeMetatable("cmd_resp_handler"))
	m.luaSelf = luaPoller
	return luaPoller
}

func registerCmdRespHandler(L *lua.LState) {
	mt := L.NewTypeMetatable("cmd_resp_handler")
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"block_get_next": pollers.PollerBlockGetNext,
		"block_has_next": pollers.PollerHasNext,
		"handle_async": func(L *lua.LState) int {
			p := checkCmdRespHandler(L)
			pollers.PollerHandleAsync(L)
			L.Push(p.luaSelf)
			return 1
		},
	}))
}

func checkCmdRespHandler(L *lua.LState) *CmdRespHandler {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*CmdRespHandler); ok {
		return v
	}
	L.ArgError(1, "cmd resp handler expected")
	return nil
}
