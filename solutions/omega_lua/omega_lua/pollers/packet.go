package pollers

import (
	"context"
	lua "github.com/yuin/gopher-lua"
)

type PacketPoller struct {
	*BasicDispatcher
	luaSelf                lua.LValue
	packetProviderCtx      context.Context
	packetProviderCancelFn func()
}

func NewPacketPoller(
	packetProviderCtx context.Context,
	packetProviderCancelFn func(),
	eventChan EventChan,
	callLua func(luaFn *lua.LFunction, numRet int, luaArgs ...lua.LValue),
) *PacketPoller {
	p := &PacketPoller{
		BasicDispatcher:        NewBasicDispatcher(eventChan, callLua),
		packetProviderCtx:      packetProviderCtx,
		packetProviderCancelFn: packetProviderCancelFn,
	}
	return p
}

func (m *PacketPoller) MakeLValue(L *lua.LState) lua.LValue {
	luaPoller := L.NewUserData()
	luaPoller.Value = m
	L.SetMetatable(luaPoller, L.GetTypeMetatable("packet_poller"))
	m.luaSelf = luaPoller
	return luaPoller
}

func registerPacketPoller(L *lua.LState) {
	mt := L.NewTypeMetatable("packet_poller")
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"stop":           packetPollerStop,
		"block_get_next": pollerBlockGetNext,
		"block_has_next": pollerHasNext,
		"handle_async": func(L *lua.LState) int {
			p := checkPacketPoller(L)
			pollerHandleAsync(L)
			L.Push(p.luaSelf)
			return 1
		},
	}))
}

func checkPacketPoller(L *lua.LState) *PacketPoller {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*PacketPoller); ok {
		return v
	}
	L.ArgError(1, "packet Poller expected")
	return nil
}

func packetPollerStop(L *lua.LState) int {
	p := checkPacketPoller(L)
	p.packetProviderCancelFn()
	return 0
}
