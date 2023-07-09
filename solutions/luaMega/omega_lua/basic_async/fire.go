package basic_async

import (
	lua "github.com/yuin/gopher-lua"
)

func FireLuaCodeInGoRoutine(ctx *AsyncCtrl, L *lua.LState, code string) (done <-chan error) {
	doneChan := make(chan error, 1)
	go func() {
		L.SetContext(ctx.Context())
		ctx.LockLState()
		err := L.DoString(code)
		ctx.UnlockLState()
		// wait until all async call finished
		ctx.WaitGroup.Wait()
		doneChan <- err
	}()
	return doneChan
}
