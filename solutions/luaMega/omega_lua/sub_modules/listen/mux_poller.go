package listen

import (
	"context"
	"fmt"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type BlockPoller struct {
	p               *OmegaBlockModule
	msgChan         chan lua.LValue
	sourceCount     int
	luaSelf         lua.LValue
	nextEvent       lua.LValue
	pCtx            context.Context
	ctx             context.Context
	cancelFn        func()
	luaFnEventAfter *lua.LFunction
}

func MakeBlockMsg(source lua.LValue, eventData map[string]lua.LValue, L *lua.LState) lua.LValue {
	event := L.NewTable()
	L.SetTable(event, lua.LString("type"), source)
	for attr, val := range eventData {
		L.SetTable(event, lua.LString(attr), val)
	}
	return event
}

func newBlockPoller(p *OmegaBlockModule, L *lua.LState, ctx context.Context) *BlockPoller {
	pollerCtx, cancelFn := context.WithCancel(ctx)
	l := &BlockPoller{
		p:        p,
		msgChan:  make(chan lua.LValue, 128),
		pCtx:     ctx,
		ctx:      pollerCtx,
		cancelFn: cancelFn,
	}
	return l
}

func (m *BlockPoller) increaseSource() {
	if m.sourceCount == 0 {
		m.ctx, m.cancelFn = context.WithCancel(m.pCtx)
	}
	m.sourceCount++
}

func (m *BlockPoller) decreaseSource() {
	m.sourceCount--
	if m.sourceCount == 0 {
		m.cancelFn()
	}
}

func (m *BlockPoller) MakeLValue(L *lua.LState) lua.LValue {
	luaPoller := L.NewUserData()
	luaPoller.Value = m
	mt := L.GetTypeMetatable("block_poller")
	eventAfter := mt.(*lua.LTable).RawGetString("__index").(*lua.LTable).RawGetString("event_after")
	m.luaFnEventAfter = eventAfter.(*lua.LFunction)
	L.SetMetatable(luaPoller, mt)
	m.luaSelf = luaPoller
	return luaPoller
}

func registerBlockPoller(L *lua.LState) {
	mt := L.NewTypeMetatable("block_poller")
	// methods
	lt := L.NewTable()
	L.SetFuncs(lt, map[string]lua.LGFunction{
		"poll":           blockPollerPoll,
		"block_get_next": blockPollerBlockGetNext,
		"has_next":       blockPollerHasNext,
		"as_async":       blockPollerAsAsync,
	})
	L.SetField(lt, "event_after", L.NewFunction(blockPollerEventAfter))
	L.SetField(mt, "__index", lt)
}

func blockPollerEventAfter(L *lua.LState) int {
	m := checkBlockPoller(L)
	timeOut := L.ToNumber(2)
	eventData := L.Get(3)
	eventDataTable := map[string]lua.LValue{"data": eventData}
	go m.handleEventAfter(float64(timeOut), m.luaFnEventAfter, eventDataTable, L)
	return 0
}

func checkBlockPoller(L *lua.LState) *BlockPoller {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*BlockPoller); ok {
		return v
	}
	L.ArgError(1, "block Poller expected")
	return nil
}

// mux_poller:poll(fn:func, args:func_args)
func blockPollerPoll(L *lua.LState) int {
	m := checkBlockPoller(L)
	target := L.Get(2)
	var err error
	switch target.Type() {
	case lua.LTFunction:
		tf := target.(*lua.LFunction)
		err = m.pollFunction(tf, L)
	case lua.LTUserData:
		err = m.pollUserData(target.(*lua.LUserData), L)
	default:
		err = fmt.Errorf("poll type not supported")
	}
	if err != nil {
		L.RaiseError(err.Error())
	}
	L.Push(m.luaSelf)
	return 1
}

// func getResArgs(L *lua.LState, fromN int) lua.LValue {
// 	numArgs := L.GetTop()
// 	if numArgs >= fromN {
// 		res := L.NewTable()
// 		for i := fromN; i <= numArgs; i++ {
// 			arg := L.Get(i)
// 			res.Append(arg)
// 		}
// 		return res
// 	}
// 	return nil
// }

func (m *BlockPoller) pollFunction(fn *lua.LFunction, L *lua.LState) error {
	if fn == m.luaFnEventAfter {
		m.increaseSource()
		timeOut := L.ToNumber(3)
		eventData := L.Get(4)
		eventDataTable := map[string]lua.LValue{"data": eventData}
		go m.handleEventAfter(float64(timeOut), m.luaFnEventAfter, eventDataTable, L)
		return nil
	}
	if fnName, found := m.p.luaFns[fn]; found {
		if fnName == LuaListenFnNameGetUserInput {
			m.increaseSource()
			go m.listenUserInput(fn, L)
		} else if fnName == LuaListenFnNameSleep {
			m.increaseSource()
			delay := L.ToNumber(3)
			go m.listenSleep(fn, float64(delay), L)
		} else {
			return fmt.Errorf("poll function not supported")
		}
	} else {
		return fmt.Errorf("poll function not supported")
	}
	return nil
}

func (m *BlockPoller) handleEventAfter(timeOut float64, event lua.LValue, eventDataTable map[string]lua.LValue, L *lua.LState) {
	defer m.decreaseSource()
	select {
	case <-m.ctx.Done():
		return
	case <-time.After(time.Duration(float64(timeOut) * float64(time.Second))):
		event := MakeBlockMsg(event, eventDataTable, L)
		m.msgChan <- event
	}
}

func (m *BlockPoller) pollUserData(ud *lua.LUserData, L *lua.LState) error {
	switch pt := ud.Value.(type) {
	case *PacketPoller:
		m.increaseSource()
		go func() {
			defer m.decreaseSource()
			for {
				pk, err := pt.GoBlockGetNext()
				if err != nil {
					break
				}
				gamePk := m.p.luaPacketModule.WrapPacketToLuaPacket(pk)
				event := MakeBlockMsg(ud, map[string]lua.LValue{
					"data": gamePk.MakeLValue(L),
				}, L)
				m.msgChan <- event
			}
		}()
	default:
		return fmt.Errorf("poll user-data not supported")
	}
	return nil
}

func (m *BlockPoller) readyNext() error {
	if m.nextEvent != nil {
		return nil
	}
	if m.sourceCount == 0 {
		return fmt.Errorf("no source")
	}
	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	case m.nextEvent = <-m.msgChan:
		return nil
	}
}

// mux_poller:block_get_next() -> event,data
func blockPollerBlockGetNext(L *lua.LState) int {
	m := checkBlockPoller(L)
	if m.readyNext() != nil {
		return 0
	}
	event := m.nextEvent
	m.nextEvent = nil
	L.Push(event)
	return 1
}

// mux_poller:has_next() -> has_next:boolean
func blockPollerHasNext(L *lua.LState) int {
	m := checkBlockPoller(L)
	L.Push(lua.LBool(m.readyNext() == nil))
	return 1
}

// mux_poller:as_async(fn:func)
func blockPollerAsAsync(L *lua.LState) int {
	m := checkBlockPoller(L)
	handler := L.ToFunction(2)
	m.p.ac.NewGoRoutine(func() {
		for {
			if err := m.readyNext(); err != nil {
				// readyNext() will return error only when ctx is done,
				// and in some cases, this is because all lua codes are done and LState is closed.
				// this is not an error, but a normal case.
				// when this happens, we don't need to raise error.
				if m.ctx.Err() == nil { // ctx (all lua code) is not done
					L.RaiseError(err.Error())
				}
				break
			}
			event := m.nextEvent
			m.nextEvent = nil
			if err := m.p.ac.SafeCall(L, lua.P{
				Fn:      handler,
				NRet:    0,
				Protect: true,
			}, event); err != nil {
				L.RaiseError(err.Error())
			}
		}
	})
	L.Push(m.luaSelf)
	return 1
}
