package pollers

import (
	"context"

	lua "github.com/yuin/gopher-lua"
)

type EventData lua.LValue
type Event lua.LValue

//struct {
//	EventType lua.LValue
//	EventData EventData
//}

//func (e *Event) MakeLValue(L *lua.LState) lua.LValue {
//	lt := L.NewTable()
//	L.SetTable(lt, lua.LString("type"), e.EventType)
//	L.SetTable(lt, lua.LString("data"), e.EventData)
//	return lt
//}

func MakeEvent(eventType lua.LValue, eventData EventData, L *lua.LState) Event {
	lt := L.NewTable()
	L.SetTable(lt, lua.LString("type"), eventType)
	L.SetTable(lt, lua.LString("data"), eventData)
	return lt
	//return &Event{
	//	EventType: eventType,
	//	EventData: eventData,
	//}
}

// EventDataChan is a channel which pass event data from source (feeder) to destination (reciver)
// feeder should assume the chan is blocked and close the chan when no more event will be produced
type EventDataChan chan lua.LValue
type EventChan chan lua.LValue

// EventSource 实现了 GetEventSource 接口
type EventDataProvider interface {
	GetFeeder() EventDataChan
}

type EventPoller interface {
	EventDataProvider
}

type LuaAsyncInvoker interface {
	SafeCall(L *lua.LState, cp lua.P, args ...lua.LValue) (err error)
	Coro(fn func())
}

type LuaEventDataChanMaker func(ctx context.Context, L *lua.LState, luaArgs ...lua.LValue) (EventDataChan, error)

func WrapEventDataChanAsDispatcher(L *lua.LState, luaAsyncInvoker LuaAsyncInvoker, maker LuaEventDataChanMaker, luaArgs ...lua.LValue) *BasicDispatcher {
	dc, err := maker(L.Context(), L, luaArgs...)
	if err != nil {
		L.RaiseError(err.Error())
	}
	c := EventChan(dc)
	//go func() {
	//	defer close(c)
	//	for {
	//		select {
	//		case <-L.Context().Done():
	//			return
	//		case eventData := <-dc:
	//			if eventData == nil {
	//				return
	//			}
	//			c <- eventData
	//		}
	//	}
	//}()
	return NewBasicDispatcher(c, luaAsyncInvoker)
}
