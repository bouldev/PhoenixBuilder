package system

import (
	"context"
	"os"
	"phoenixbuilder/solutions/omega_lua/omega_lua/concurrent"
	"runtime"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// system module os/sleep/cwd/set_timeout/set_interval/now
type OmegaSystemModule struct {
	ac        *concurrent.AsyncCtrl
	startTime time.Time
}

func NewOmegaSystemModule(ac *concurrent.AsyncCtrl) *OmegaSystemModule {
	return &OmegaSystemModule{
		ac:        ac,
		startTime: time.Now(),
	}
}

func (m *OmegaSystemModule) MakeLValue(L *lua.LState) lua.LValue {
	return L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"os":           m.luaGoSystemOs,
		"cwd":          m.luaGoSystemCwd,
		"now":          m.luaGoSystemNow,
		"set_timeout":  m.luaGoSetTimeout,
		"set_interval": m.luaGoSetTimeInterval,
	})
}

// system.os() arch-system:string
func (m *OmegaSystemModule) luaGoSystemOs(L *lua.LState) int {
	osName := runtime.GOOS
	archName := runtime.GOARCH
	L.Push(lua.LString(osName + "-" + archName))
	return 1
}

// system.cwd() cwd:string
func (m *OmegaSystemModule) luaGoSystemCwd(L *lua.LState) int {
	cwd, err := os.Getwd()
	if err != nil {
		L.RaiseError(err.Error())
		return 0
	}
	L.Push(lua.LString(cwd))
	return 1
}

// system.now() now_second:number
func (m *OmegaSystemModule) luaGoSystemNow(L *lua.LState) int {
	L.Push(lua.LNumber(float64(time.Since(m.startTime).Seconds())))
	return 1
}

// system.set_timeout(second:number, fn:func)
func (m *OmegaSystemModule) luaGoSetTimeout(L *lua.LState) int {
	timeout := time.Duration(float64(L.ToNumber(1)) * float64(time.Second))
	callback := L.ToFunction(2)
	m.ac.NewGoRoutine(func() {
		time.Sleep(timeout)
		err := m.ac.SafeCall(L, lua.P{
			Fn:      callback,
			NRet:    0,
			Protect: true,
		})
		if err != nil {
			L.RaiseError(err.Error())
		}
	})
	return 0
}

// system.set_interval(second:number, fn:func)
func (m *OmegaSystemModule) luaGoSetTimeInterval(L *lua.LState) int {
	interval := time.Duration(float64(L.ToNumber(1)) * float64(time.Second))
	callback := L.ToFunction(2)
	stopCtx, stopFn := context.WithCancel(m.ac.Context())
	go func() {
		<-m.ac.Context().Done()
		stopFn()
	}()
	m.ac.NewGoRoutine(func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				err := m.ac.SafeCall(L, lua.P{
					Fn:      callback,
					NRet:    0,
					Protect: true,
				})
				if err != nil {
					L.RaiseError(err.Error())
				}
			case <-stopCtx.Done():
				// ticker.Stop()
				return
			}
		}
	})
	luaStopFn := L.NewFunction(func(L *lua.LState) int {
		stopFn()
		return 0
	})
	L.Push(luaStopFn)
	return 1
}
