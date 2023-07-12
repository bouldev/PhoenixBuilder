package pollers

import (
	lua "github.com/yuin/gopher-lua"
)

type EventData lua.LValue
type Event EventData

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
type EventDataChan chan EventData
type EventChan chan Event

// EventSource 实现了 GetEventSource 接口
type EventDataProvider interface {
	GetFeeder() EventDataChan
}

type EventPoller interface {
	EventDataProvider
}

type LuaInvoker interface {
	CallLua(luaFn *lua.LFunction, numRet int, luaArgs ...lua.LValue)
	AddCoro(int)
	DecreaseCoro()
}
