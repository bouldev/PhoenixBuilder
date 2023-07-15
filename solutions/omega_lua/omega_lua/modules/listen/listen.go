package listen

import (
	"context"
	"phoenixbuilder/minecraft/protocol/packet"
	"phoenixbuilder/solutions/omega_lua/omega_lua/pollers"

	lua "github.com/yuin/gopher-lua"
)

type PacketProvider interface {
	MakeMCPacketFeeder(ctx context.Context, wants []string) <-chan packet.Packet
}

type ListenModule struct {
	pollers.LuaAsyncInvoker
	goPacketsProvider PacketProvider
	packetWrapper     func(goPacket packet.Packet, L *lua.LState) lua.LValue
	muxPollerFlags    map[lua.LValue]pollers.LuaEventDataChanMaker
}

func NewListenModule(
	luaAsyncInvoker pollers.LuaAsyncInvoker,
	goPacketsProvider PacketProvider,
	packetWrapper func(goPacket packet.Packet, L *lua.LState) lua.LValue,
	flagGroup ...map[lua.LValue]pollers.LuaEventDataChanMaker,
) *ListenModule {
	flags := make(map[lua.LValue]pollers.LuaEventDataChanMaker)
	for _, flag := range flagGroup {
		for k, v := range flag {
			flags[k] = v
		}
	}
	m := &ListenModule{
		LuaAsyncInvoker:   luaAsyncInvoker,
		goPacketsProvider: goPacketsProvider,
		packetWrapper:     packetWrapper,
		muxPollerFlags:    flags,
	}
	return m
}

func (m *ListenModule) MakeLValue(L *lua.LState) lua.LValue {
	luaModule := L.NewTable()
	luaModule = L.SetFuncs(luaModule, map[string]lua.LGFunction{
		"new_mux_poller":    m.luaGoMakeMuxPoller,
		"new_packet_poller": m.luaGoMakePacketPoller,
	})
	pollers.RegisterBlockPoller(L)
	pollers.RegisterPacketPoller(L)
	return luaModule
}

func (m *ListenModule) luaGoMakeMuxPoller(L *lua.LState) int {
	mux := pollers.NewBlockPoller(m.muxPollerFlags, m.LuaAsyncInvoker, L.Context())
	mux_poller := mux.MakeLValue(L)
	L.Push(mux_poller)
	return 1
}

func (m *ListenModule) luaGoMakePacketPoller(L *lua.LState) int {
	packetNames := make([]string, L.GetTop())
	for i := 1; i <= L.GetTop(); i++ {
		packetNames[i-1] = L.CheckString(i)
	}
	ctx, cancelFn := context.WithCancel(L.Context())
	packetChan := m.goPacketsProvider.MakeMCPacketFeeder(ctx, packetNames)
	eventChan := make(chan lua.LValue)
	go func() {
		defer func() {
			close(eventChan)
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case p := <-packetChan:
				if p == nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case eventChan <- m.packetWrapper(p, L):
				}
			}
		}
	}()
	poller := pollers.NewPacketPoller(ctx, cancelFn, eventChan, m.LuaAsyncInvoker)
	luaPoller := poller.MakeLValue(L)
	L.Push(luaPoller)
	return 1
}
