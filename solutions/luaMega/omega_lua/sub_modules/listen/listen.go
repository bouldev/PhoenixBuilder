package listen

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

const (
	LuaListenFnNameSleep        = "sleep"
	LuaListenFnNameGetUserInput = "user_input"
)

// block.user_input() input:string
func (m *OmegaBlockModule) luaGoBlockGetUserInput(L *lua.LState) int {
	userInput := <-m.goImplements.UserInputChan()
	L.Push(lua.LString(userInput))
	return 1
}

// poller.poll.user_input -> input:string
func (m *BlockPoller) listenUserInput(fn lua.LValue, L *lua.LState) {
	defer m.decreaseSource()
	select {
	case <-m.ctx.Done():
		return
	case input := <-m.p.goImplements.UserInputChan():
		event := MakeBlockMsg(fn, map[string]lua.LValue{
			"data": lua.LString(input),
		}, L)
		m.msgChan <- event
	}
}

// block module: sleep(second:number)
func (m *OmegaBlockModule) luaGoBlockSleep(L *lua.LState) int {
	time.Sleep(time.Duration(float64(L.ToNumber(1)) * float64(time.Second)))
	return 0
}

// listen:sleep() -> nil
func (m *BlockPoller) listenSleep(fn lua.LValue, delay float64, L *lua.LState) {
	defer m.decreaseSource()
	select {
	case <-m.ctx.Done():
		return
	case <-time.After(time.Duration(float64(delay) * float64(time.Second))):
		event := MakeBlockMsg(fn, map[string]lua.LValue{}, L)
		m.msgChan <- event
	}
}
