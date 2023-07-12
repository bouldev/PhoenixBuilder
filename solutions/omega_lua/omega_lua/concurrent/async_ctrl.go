package concurrent

import (
	"context"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// 异步管理器
type AsyncCtrl struct {
	*sync.WaitGroup
	closed   bool
	ctx      context.Context
	cancelFn func()
	mu       sync.Mutex
}

// 获取一个新的异步管理器
func NewAsyncCtrl(baseCtx context.Context) *AsyncCtrl {
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	ctx, cancelFn := context.WithCancel(baseCtx)

	a := &AsyncCtrl{
		WaitGroup: &sync.WaitGroup{},
		closed:    false,
		ctx:       ctx,
		cancelFn:  cancelFn,
		mu:        sync.Mutex{},
	}
	go func() {
		<-ctx.Done()    // 监听 ctx 的完成事件
		a.closed = true // 标记异步管理器已关闭
	}()
	return a
}

func (a *AsyncCtrl) IsClosed() bool {
	return a.closed
}

// 取得上下文
func (a *AsyncCtrl) Context() context.Context {
	return a.ctx
}

// 锁住解释器
// The LState is not goroutine-safe. It is recommended to use one LState per goroutine and communicate between goroutines by using channels.
func (a *AsyncCtrl) LockLState() {
	a.mu.Lock()
}

// 解锁解释器
func (a *AsyncCtrl) UnlockLState() {
	a.mu.Unlock()
}

// 安全地调用lua函数 第一个参数为解释器 第二个参数为函数 第三个参数为函数的参数
func (a *AsyncCtrl) SafeCall(L *lua.LState, cp lua.P, args ...lua.LValue) (err error) {
	if a.IsClosed() {
		return nil
	}
	a.LockLState()
	defer a.UnlockLState()
	return L.CallByParam(cp, args...)
}

// 创建一个新的go线程并且请求完成指定函数
// 必须先完成
// 如果不完成则堵塞
func (a *AsyncCtrl) NewGoRoutine(fn func()) {
	if a.IsClosed() {
		return
	}
	a.Add(1) // 添加计数器
	go func() {
		defer a.Done() // 减少计数器
		fn()           // 执行传入的函数
	}()
}
