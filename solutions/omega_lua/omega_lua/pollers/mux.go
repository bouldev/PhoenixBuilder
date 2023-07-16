package pollers

import (
	"context"
	"errors"

	lua "github.com/yuin/gopher-lua"
)

type BasicMux struct {
	*BasicDispatcher
	LuaAsyncInvoker
	parentCtx   context.Context
	eventChan   EventChan
	ctx         context.Context
	cancelFn    func()
	sourceCount int
	// reserveOnNoSource: if true, eventChan will not be closed when no source
	reserveOnNoSource bool
}

func NewBasicMux(parentCtx context.Context, luaAsyncInvoker LuaAsyncInvoker) *BasicMux {
	eventChan := make(chan lua.LValue)
	ctx, cancelFn := context.WithCancel(parentCtx)
	m := &BasicMux{
		BasicDispatcher:   NewBasicDispatcher(eventChan, luaAsyncInvoker),
		LuaAsyncInvoker:   luaAsyncInvoker,
		eventChan:         eventChan,
		parentCtx:         parentCtx,
		ctx:               ctx,
		cancelFn:          cancelFn,
		reserveOnNoSource: true,
	}
	return m
}

func (m *BasicMux) SetReserve(b bool) {
	if m.reserveOnNoSource == b {
		return
	}
	if m.sourceCount == 0 {
		if b {
			// reserve
			m.eventChan = make(chan lua.LValue)
		} else {
			// unreserve
			close(m.eventChan)
			// if m.ctx.Err() == nil {
			// 	m.cancelFn()
			// }
		}
	}
	m.reserveOnNoSource = b
}

func (m *BasicMux) increaseSource() {
	if m.sourceCount == 0 {
		m.ctx, m.cancelFn = context.WithCancel(m.parentCtx) // new source
		if !m.reserveOnNoSource {
			m.eventChan = make(chan lua.LValue)
		}
	}
	m.sourceCount++
	// fmt.Printf("increase source %v,\n", m.sourceCount)
}

func (m *BasicMux) decreaseSource() {
	m.sourceCount--
	// fmt.Printf("decrease source %v,\n", m.sourceCount)
	if m.sourceCount == 0 {
		m.cancelFn()
		if !m.reserveOnNoSource {
			close(m.eventChan)
		}
	}
}

func muxEventChan(ctx context.Context, subChan EventDataChan, eventType lua.LValue, muxChan EventChan, L *lua.LState) {
	for {
		select {
		case <-ctx.Done():
			return
		case eventData := <-subChan:
			if eventData == nil {
				return
			}
			event := MakeEvent(eventType, eventData, L)
			muxChan <- event
		}
	}

}

func (m *BasicMux) AddSource(subChan EventDataChan, eventType lua.LValue, L *lua.LState) {
	m.Coro(func() {
		muxEventChan(m.ctx, subChan, eventType, m.eventChan, L)
		m.decreaseSource()
	})
	m.increaseSource()
}

func (m *BasicMux) AddProvider(provider EventDataProvider, eventType lua.LValue, L *lua.LState) {
	m.AddSource(provider.GetFeeder(), eventType, L)
}

type BasicDispatcher struct {
	eventChan EventChan
	nextEvent Event
	LuaAsyncInvoker
}

// when eventChan closed, all dispatch will stop
func NewBasicDispatcher(eventChan EventChan, luaAsyncInvoker LuaAsyncInvoker) *BasicDispatcher {
	return &BasicDispatcher{
		LuaAsyncInvoker: luaAsyncInvoker,
		eventChan:       eventChan,
	}
}

func (d *BasicDispatcher) GetFeeder() EventDataChan {
	//eventDataChan := make(chan EventData)
	//go func() {
	//	defer close(eventDataChan)
	//	for {
	//		//select {
	//		//case <-d.ctx.Done():
	//		//	break
	//		//case
	//		event := <-d.eventChan
	//		if event == nil {
	//			break
	//		}
	//		eventDataChan <- EventData(event)
	//		//}
	//	}
	//}()
	//return eventDataChan
	return EventDataChan(d.eventChan)
}

var ErrNoMoreEvent = errors.New("no more events")

// wait and put next event to m.next, error is return if context is done
func (d *BasicDispatcher) blockReadyNext() error {
	if d.nextEvent != nil {
		return nil
	}
	//select {
	//case <-d.ctx.Done():
	//	return d.ctx.Err()
	//case d.nextEvent = <-d.eventChan:
	//	return nil
	//}
	d.nextEvent = <-d.eventChan
	if d.nextEvent == nil {
		return ErrNoMoreEvent
	}
	return nil
}

func (d *BasicDispatcher) blockHasNext() bool {
	return d.blockReadyNext() == nil
}

func (d *BasicDispatcher) blockGetNext() Event {
	if d.blockReadyNext() != nil {
		return nil
	}
	event := d.nextEvent
	d.nextEvent = nil
	return event
}

func (d *BasicDispatcher) SetHandler(cb func(event Event)) {
	d.Coro(func() {
		for {
			event := d.blockGetNext()
			if event == nil {
				break
			}
			//if d.ctx.Err() != nil {
			//	break
			//}
			cb(event)
		}
	})
}

func (d *BasicDispatcher) LuaBlockHasNext(L *lua.LState) int {
	L.Push(lua.LBool(d.blockHasNext()))
	return 1
}

func (d *BasicDispatcher) LuaBlockGetNext(L *lua.LState) int {
	if event := d.blockGetNext(); event == nil {
		return 0
	} else {
		L.Push(event)
		return 1
	}
}

type CanSetReserve interface {
	SetReserve(b bool)
}

func checkCanSetReserve(L *lua.LState) CanSetReserve {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(CanSetReserve); ok {
		return v
	}
	L.ArgError(1, "poller expected")
	return nil
}

func pollerReserve(L *lua.LState) int {
	m := checkCanSetReserve(L)
	m.SetReserve(L.ToBool(2))
	return 0
}

type CanBlockGetNext interface {
	LuaBlockGetNext(L *lua.LState) int
}

func checkCanBlockGetNext(L *lua.LState) CanBlockGetNext {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(CanBlockGetNext); ok {
		return v
	}
	L.ArgError(1, "poller expected")
	return nil
}

func PollerBlockGetNext(L *lua.LState) int {
	m := checkCanBlockGetNext(L)
	return m.LuaBlockGetNext(L)
}

type CanBlockHasNext interface {
	LuaBlockHasNext(L *lua.LState) int
}

func checkCanBlockHasNext(L *lua.LState) CanBlockHasNext {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(CanBlockHasNext); ok {
		return v
	}
	L.ArgError(1, "poller expected")
	return nil
}

func PollerHasNext(L *lua.LState) int {
	m := checkCanBlockHasNext(L)
	return m.LuaBlockHasNext(L)
}

type CanSetHandler interface {
	SetHandler(cb func(event Event))
}

type CanSetLuaHandler interface {
	SetHandler(cb func(event Event))
	LuaAsyncInvoker
}

func checkCanSetLuaHandler(L *lua.LState) CanSetLuaHandler {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(CanSetLuaHandler); ok {
		return v
	}
	L.ArgError(1, "poller expected")
	return nil
}

func PollerHandleAsync(L *lua.LState) int {
	m := checkCanSetLuaHandler(L)
	handler := L.ToFunction(2)
	m.SetHandler(func(event Event) {
		m.SafeCall(L, lua.P{
			Fn:      handler,
			NRet:    0,
			Protect: false,
		}, event)
	})
	return 0
}
