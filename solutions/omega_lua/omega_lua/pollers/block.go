package pollers

import (
	"context"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type BlockPoller struct {
	*BasicMux
	// poller:poll(flag,args...) self
	eventMakersOnFlag  map[lua.LValue]LuaEventDataChanMaker
	luaSelf            lua.LValue
	selfFlagEventAfter lua.LValue
}

func RegisterBlockPoller(L *lua.LState) {
	mt := L.NewTypeMetatable("block_poller")
	// methods
	lt := L.NewTable()
	L.SetFuncs(lt, map[string]lua.LGFunction{
		"poll":           blockPollerPoll,
		"stop":           blockPollerStop,
		"block_get_next": PollerBlockGetNext,
		"block_has_next": PollerHasNext,
		"reserve":        pollerReserve,
		"handle_async": func(L *lua.LState) int {
			p := checkPacketPoller(L)
			PollerHandleAsync(L)
			L.Push(p.luaSelf)
			return 1
		},
	})
	L.SetField(lt, "event_after", L.NewFunction(blockPollerEventAfter))
	L.SetField(mt, "__index", lt)
}

func NewBlockPoller(
	outerFlags map[lua.LValue]LuaEventDataChanMaker,
	luaAsyncInvoker LuaAsyncInvoker,
	ctx context.Context,
) *BlockPoller {
	if outerFlags == nil {
		outerFlags = map[lua.LValue]LuaEventDataChanMaker{}
	}
	p := &BlockPoller{
		BasicMux:          NewBasicMux(ctx, luaAsyncInvoker),
		eventMakersOnFlag: outerFlags,
	}
	return p
}

func (m *BlockPoller) MakeLValue(L *lua.LState) lua.LValue {
	luaPoller := L.NewUserData()
	luaPoller.Value = m
	mt := L.GetTypeMetatable("block_poller")
	{ // inject support for poller:poll(poller.event_after,time,data)
		m.selfFlagEventAfter = mt.(*lua.LTable).RawGetString("__index").(*lua.LTable).RawGetString("event_after")
		m.eventMakersOnFlag[m.selfFlagEventAfter] = m.eventAfterSourceMaker
	}
	L.SetMetatable(luaPoller, mt)
	m.luaSelf = luaPoller
	return luaPoller
}

func (m *BlockPoller) eventAfterSourceMaker(ctx context.Context, L *lua.LState, luaArgs ...lua.LValue) (EventDataChan, error) {
	if len(luaArgs) < 2 {
		return nil, fmt.Errorf("in sufficient args number")
	}
	timeOut := float64(lua.LVAsNumber(luaArgs[0]))
	eventData := luaArgs[1]
	c := make(chan lua.LValue)
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(float64(timeOut) * float64(time.Second))):
			c <- eventData
		}
		close(c)
	}()
	return c, nil
}

func (m *BlockPoller) addSourceByEventFlagAndLuaArgs(flag lua.LValue, L *lua.LState, luaArgs ...lua.LValue) error {
	maker, found := m.eventMakersOnFlag[flag]
	if !found {
		return fmt.Errorf("poll flag not found")
	}
	newSource, err := maker(m.ctx, L, luaArgs...)
	if err != nil {
		return err
	}
	m.AddSource(newSource, flag, L)
	return nil
}

func checkBlockPoller(L *lua.LState) *BlockPoller {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*BlockPoller); ok {
		return v
	}
	L.ArgError(1, "block poller expected")
	return nil
}

func blockPollerStop(L *lua.LState) int {
	m := checkBlockPoller(L)
	m.cancelFn()
	close(m.eventChan)
	L.Push(m.luaSelf)
	return 1
}

func blockPollerEventAfter(L *lua.LState) int {
	m := checkBlockPoller(L)
	err := m.addSourceByEventFlagAndLuaArgs(m.selfFlagEventAfter, L, L.Get(2), L.Get(3))
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	L.Push(m.luaSelf)
	return 1
}

func blockPollerPoll(L *lua.LState) int {
	m := checkBlockPoller(L)
	eventType := L.Get(2)
	var err error
	switch eventType.Type() {
	case lua.LTFunction:
		resArgs := make([]lua.LValue, 0)
		for i := 3; i <= L.GetTop(); i++ {
			resArgs = append(resArgs, L.Get(i))
		}
		err = m.addSourceByEventFlagAndLuaArgs(eventType, L, resArgs...)
	case lua.LTUserData:
		couldBeProvider := eventType.(*lua.LUserData).Value
		if provider, ok := couldBeProvider.(EventDataProvider); ok {
			m.AddProvider(provider, eventType, L)
		} else {
			err = fmt.Errorf("not a pollable object")
		}
	default:
		err = fmt.Errorf("poll type not supported")
	}
	if err != nil {
		panic(err)
		L.ArgError(1, err.Error())
	}
	L.Push(m.luaSelf)
	return 1
}
