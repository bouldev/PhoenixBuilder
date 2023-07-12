package pollers

import (
	"context"
	"errors"
	lua "github.com/yuin/gopher-lua"
)

func muxEventChan(ctx context.Context, subChan EventDataChan, eventType lua.LValue, muxChan EventChan, L *lua.LState) {
	for {
		select {
		case <-ctx.Done():
			break
		case eventData := <-subChan:
			if eventData == nil {
				break
			}
			event := MakeEvent(eventType, eventData, L)
			muxChan <- event
		}
	}

}

type BasicMux struct {
	*BasicDispatcher
	parentCtx   context.Context
	eventChan   EventChan
	ctx         context.Context
	cancelFn    func()
	sourceCount int
}

func NewBasicMux(ctx context.Context, callLua func(luaFn *lua.LFunction, numRet int, luaArgs ...lua.LValue)) *BasicMux {
	eventChan := make(chan Event)
	m := &BasicMux{
		parentCtx:       ctx,
		BasicDispatcher: NewBasicDispatcher(eventChan, callLua),
	}
	return m
}

// this block has next has different behaviour than dispatcher
// where the dispather dicide has next only on the open statues of chan
// but mux dicided on also whether there is a source
func (m *BasicMux) LuaBlockHasNext(L *lua.LState) int {
	hasNext := false
	if m.BasicDispatcher.nextEvent != nil {
		hasNext = true
	} else {
		select {
		case <-m.ctx.Done(): // no source
		case m.BasicDispatcher.nextEvent = <-m.BasicDispatcher.eventChan:
			hasNext = true
		}
	}
	L.Push(lua.LBool(hasNext))
	return 1
}

func (m *BasicMux) increaseSource() {
	if m.sourceCount == 0 {
		m.ctx, m.cancelFn = context.WithCancel(m.parentCtx) // new source
	}
	m.sourceCount++
}

func (m *BasicMux) decreaseSource() {
	m.sourceCount--
	if m.sourceCount == 0 {
		m.cancelFn()
		//close(m.eventChan) // could add source later so don't close it
		//m.ctx, m.cancelFn = nil, nil
	}
}

func (m *BasicMux) AddSource(subChan EventDataChan, eventType lua.LValue, L *lua.LState) {
	go func() {
		muxEventChan(m.ctx, subChan, eventType, m.eventChan, L)
		m.decreaseSource()
	}()
	m.increaseSource()
}

func (m *BasicMux) AddProvider(provider EventDataProvider, eventType lua.LValue, L *lua.LState) {
	m.AddSource(provider.GetFeeder(), eventType, L)
}

type BasicDispatcher struct {
	eventChan EventChan
	nextEvent Event
	callLua   func(luaFn *lua.LFunction, numRet int, luaArgs ...lua.LValue)
}

// when eventChan closed, all dispatch will stop
func NewBasicDispatcher(eventChan EventChan, callLua func(luaFn *lua.LFunction, numRet int, luaArgs ...lua.LValue)) *BasicDispatcher {
	return &BasicDispatcher{
		callLua:   callLua,
		eventChan: eventChan,
	}
}

func (d *BasicDispatcher) GetFeeder() EventDataChan {
	eventDataChan := make(chan EventData)
	go func() {
		defer close(eventDataChan)
		for {
			//select {
			//case <-d.ctx.Done():
			//	break
			//case
			event := <-d.eventChan
			if event == nil {
				break
			}
			eventDataChan <- EventData(event)
			//}
		}
	}()
	return eventDataChan
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

type CanBlockGetNext interface {
	LuaBlockGetNext(L *lua.LState) int
}

type CanBlockHasNext interface {
	LuaBlockHasNext(L *lua.LState) int
}

type CanSetHandler interface {
	SetHandler(cb func(event Event))
}

type CanSetLuaHandler interface {
	SetHandler(cb func(event Event))
	CallLua(luaFn *lua.LFunction, numRet int, luaArgs ...lua.LValue)
}

func checkCanBlockGetNext(L *lua.LState) CanBlockGetNext {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(CanBlockGetNext); ok {
		return v
	}
	L.ArgError(1, "poller expected")
	return nil
}

func checkCanBlockHasNext(L *lua.LState) CanBlockHasNext {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(CanBlockHasNext); ok {
		return v
	}
	L.ArgError(1, "poller expected")
	return nil
}

func checkCanSetLuaHandler(L *lua.LState) CanSetLuaHandler {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(CanSetLuaHandler); ok {
		return v
	}
	L.ArgError(1, "poller expected")
	return nil
}

func pollerBlockGetNext(L *lua.LState) int {
	m := checkCanBlockGetNext(L)
	return m.LuaBlockGetNext(L)
}

func pollerHasNext(L *lua.LState) int {
	m := checkCanBlockHasNext(L)
	return m.LuaBlockHasNext(L)
}

func pollerHandleAsync(L *lua.LState) int {
	m := checkCanSetLuaHandler(L)
	handler := L.ToFunction(2)
	m.SetHandler(func(event Event) {
		m.CallLua(handler, 0, event)
	})
	return 0
}
