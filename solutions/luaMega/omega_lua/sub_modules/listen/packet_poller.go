package listen

import (
	"context"
	"phoenixbuilder/minecraft/protocol/packet"

	lua "github.com/yuin/gopher-lua"
)

type PacketPoller struct {
	p        *OmegaBlockModule
	luaSelf  lua.LValue
	pkChan   <-chan packet.Packet
	cancelFn func()
	ctx      context.Context
	next     packet.Packet
}

func newPacketPoller(p *OmegaBlockModule,
	pkChan <-chan packet.Packet,
	ctx context.Context,
	cancelFn func()) *PacketPoller {

	return &PacketPoller{
		p:        p,
		pkChan:   pkChan,
		ctx:      ctx,
		cancelFn: cancelFn,
	}
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
		"has_next":       packetPollerHasNext,
		"block_get_next": packetPollerBlockGetNext,
		"stop":           packetPollerStop,
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

// packet_poller.has_next() -> has_next:boolean
func packetPollerHasNext(L *lua.LState) int {
	p := checkPacketPoller(L)
	b := p.GoHasNext()
	L.Push(lua.LBool(b))
	return 1
}

func (m *PacketPoller) readyNext() error {
	if m.next != nil {
		return nil
	}
	select {
	case m.next = <-m.pkChan:
		return nil
	case <-m.ctx.Done():
		return m.ctx.Err()
	}
}

func (m *PacketPoller) GoHasNext() bool {
	return m.readyNext() == nil
}

// packet_poller.block_get_next() -> packet:table
func packetPollerBlockGetNext(L *lua.LState) int {
	p := checkPacketPoller(L)
	pk, err := p.GoBlockGetNext()
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	luaPacket := p.p.luaPacketModule.WrapPacketToLuaPacket(pk)
	L.Push(luaPacket.MakeLValue(L))
	return 1
}

func (m *PacketPoller) GoBlockGetNext() (pk packet.Packet, err error) {
	err = m.readyNext()
	if err != nil {
		return nil, err
	}
	pk = m.next
	m.next = nil
	return pk, nil
}

func packetPollerStop(L *lua.LState) int {
	p := checkPacketPoller(L)
	p.cancelFn()
	return 0
}
