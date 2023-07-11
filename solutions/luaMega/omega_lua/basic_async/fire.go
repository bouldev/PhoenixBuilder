package basic_async

import (
	lua "github.com/yuin/gopher-lua"
)

// FireLuaCodeInGoRoutine 在一个新的 Go 协程中执行给定的 Lua 代码，并返回一个错误通道
// 热启动lua代码
func FireLuaCodeInGoRoutine(ctx *AsyncCtrl, L *lua.LState, code string) (done <-chan error) {
	// 创建一个带缓冲的错误通道 用于取消这个线程
	//办到中断这个lua代码
	doneChan := make(chan error, 1)
	go func() {
		// 将上下文设置到 LState 中
		L.SetContext(ctx.Context())
		// 锁住整个解释器
		ctx.LockLState()
		err := L.DoString(code)
		//解锁解释器
		ctx.UnlockLState()
		// wait until all async call finished
		//当所有的异步都加载完毕再返回取消消息
		ctx.WaitGroup.Wait()
		doneChan <- err
	}()
	return doneChan
}
