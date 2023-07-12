package concurrent

import (
	"context"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

// 异步管理器
type AsyncCtrl struct {
	*sync.WaitGroup
	closed bool
	ctx    context.Context
	mu     sync.Mutex
}

// 获取一个新的异步管理器
func NewAsyncCtrl(ctx context.Context) *AsyncCtrl {
	if ctx == nil {
		ctx = context.Background()
	}
	a := &AsyncCtrl{
		WaitGroup: &sync.WaitGroup{},
		closed:    false,
		ctx:       ctx,
		mu:        sync.Mutex{},
	}
	return a
}

// 取得上下文
func (a *AsyncCtrl) Context() context.Context {
	return a.ctx
}

// 锁住解释器
// The LState is not goroutine-safe. It is recommended to use one LState per goroutine and communicate between goroutines by using channels.
func (a *AsyncCtrl) lockLState() {
	a.mu.Lock()
}

// 解锁解释器
func (a *AsyncCtrl) unlockLState() {
	a.mu.Unlock()
}

// 安全地调用lua函数 第一个参数为解释器 第二个参数为函数 第三个参数为函数的参数
func (a *AsyncCtrl) SafeCall(L *lua.LState, cp lua.P, args ...lua.LValue) (err error) {
	if a.ctx.Err() != nil {
		return nil
	}
	a.lockLState()
	defer a.unlockLState()
	return L.CallByParam(cp, args...)
}

func (a *AsyncCtrl) AddCoro(count int) error {
	if a.ctx.Err() != nil {
		return a.ctx.Err()
	}
	a.Add(count)
	return nil
}

func (a *AsyncCtrl) DecreaseCoro() {
	a.Done()
}
