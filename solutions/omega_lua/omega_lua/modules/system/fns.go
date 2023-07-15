package system

import (
	"context"
	"os"
	"phoenixbuilder/solutions/omega_lua/omega_lua/pollers"
	"runtime"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func (m *SystemModule) luaGoSystemPrint(L *lua.LState) int {
	m.LuaGoSystem.Print(L.ToString(-1))
	return 0
}

func (m *SystemModule) luaGoSystemOs(L *lua.LState) int {
	osName := runtime.GOOS
	archName := runtime.GOARCH
	L.Push(lua.LString(osName + "-" + archName))
	return 1
}

func (m *SystemModule) luaGoSystemCwd(L *lua.LState) int {
	cwd, err := os.Getwd()
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	L.Push(lua.LString(cwd))
	return 1
}

func goSleepSourceMaker(ctx context.Context, L *lua.LState, luaArgs ...lua.LValue) (pollers.EventDataChan, error) {
	c := make(chan lua.LValue)
	timeOut := float64(lua.LVAsNumber(luaArgs[0]))
	go func() {
		select {
		case <-ctx.Done():
		case <-time.After(time.Duration(float64(timeOut) * float64(time.Second))):
			c <- lua.LNil
		}
		close(c)
	}()
	return c, nil
}

func (m *SystemModule) luaGoSleep(L *lua.LState) int {
	return pollers.WrapEventDataChanAsDispatcher(L, m.LuaAsyncInvoker, goSleepSourceMaker, L.Get(1)).
		LuaBlockGetNext(L)
}

func (m *SystemModule) goInputSourceMaker(ctx context.Context, L *lua.LState, luaArgs ...lua.LValue) (pollers.EventDataChan, error) {
	c := make(chan lua.LValue)
	go func() {
		select {
		case <-ctx.Done():
		case c <- lua.LString(<-m.LuaGoSystem.UserInputChan()):
		}
		close(c)
	}()
	return c, nil
}

func (m *SystemModule) luaGoInput(L *lua.LState) int {
	return pollers.WrapEventDataChanAsDispatcher(L, m.LuaAsyncInvoker, m.goInputSourceMaker).
		LuaBlockGetNext(L)
}
