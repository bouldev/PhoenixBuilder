package concurrent

import (
	"context"
	"sync"
	"time"

	"github.com/pterm/pterm"
	lua "github.com/yuin/gopher-lua"
)

// 异步管理器
type asyncCtrl struct {
	*sync.WaitGroup
	closed    bool
	ctx       context.Context
	mu        sync.Mutex
	coroCount int
}

type AsyncCtrl interface {
	// 使用 AddCoro 和 DecreaseCoro 标记尚在进行中的 lua 代码
	AddCoro(count int) error
	DecreaseCoro()
	Coro(fn func())
	// 在异步回调中使用 SafeCall 来保证并发安全
	SafeCall(L *lua.LState, cp lua.P, args ...lua.LValue) (err error)
	// Context 读取 上下文
	Context() context.Context
	// 等待所有尚未完成的lua代码完成
	Wait()
}

// 获取一个新的异步管理器
func NewAsyncCtrl(ctx context.Context) AsyncCtrl {
	if ctx == nil {
		ctx = context.Background()
	}
	a := &asyncCtrl{
		WaitGroup: &sync.WaitGroup{},
		closed:    false,
		ctx:       ctx,
		mu:        sync.Mutex{},
		coroCount: 0,
	}
	return a
}

// 取得上下文
func (a *asyncCtrl) Context() context.Context {
	return a.ctx
}

// 锁住解释器
// The LState is not goroutine-safe. It is recommended to use one LState per goroutine and communicate between goroutines by using channels.
func (a *asyncCtrl) lockLState() {
	a.mu.Lock()
}

// 解锁解释器
func (a *asyncCtrl) unlockLState() {
	a.mu.Unlock()
}

// 安全地调用lua函数 第一个参数为解释器 第二个参数为函数 第三个参数为函数的参数
func (a *asyncCtrl) SafeCall(L *lua.LState, cp lua.P, args ...lua.LValue) (err error) {
	if a.ctx.Err() != nil {
		return nil
	}
	startTime := time.Now()
	a.lockLState()
	defer a.unlockLState()
	if time.Since(startTime) > time.Second {
		pterm.Warning.Printf("一个异步回调函数被调用的时间比设定好的显著延迟了，您是否混用了异步回调(async)的代码和同步阻塞(block)的代码？这会导致回调函数无法正常工作")
	}
	return L.CallByParam(cp, args...)
}

func (a *asyncCtrl) AddCoro(count int) error {
	if a.ctx.Err() != nil {
		return a.ctx.Err()
	}
	a.Add(count)
	a.coroCount += count
	return nil
}

func (a *asyncCtrl) DecreaseCoro() {
	a.coroCount--
	a.Done()
}

func (a *asyncCtrl) Coro(fn func()) {
	a.Add(1)
	go func() {
		defer a.Done()
		fn()
	}()
}
