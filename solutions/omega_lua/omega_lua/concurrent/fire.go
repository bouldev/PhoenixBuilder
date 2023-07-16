package concurrent

import (
	lua "github.com/yuin/gopher-lua"
)

// FireLuaCodeInGoRoutine 在一个新的 Go 协程中执行给定的 Lua 代码，并返回一个错误通道
// 热启动lua代码
func FireLuaCodeInGoRoutine(ac AsyncCtrl, L *lua.LState, code string) (done <-chan error) {
	// 创建一个带缓冲的错误通道 用于取消这个线程
	//办到中断这个lua代码
	doneChan := make(chan error)
	var err error
	ac.AddCoro(1)
	go func() {
		defer ac.DecreaseCoro()
		L.SetContext(ac.Context())
		err = L.DoString(code)
		if err != nil {
			doneChan <- err
		}
	}()
	go func() {
		ac.Wait()
		close(doneChan)
	}()
	return doneChan
}
