package basic_async

import (
	"context"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

type AsyncCtrl struct {
	*sync.WaitGroup
	closed   bool
	ctx      context.Context
	cancelFn func()
	mu       sync.Mutex
}

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
		<-ctx.Done()
		a.closed = true
	}()
	return a
}

func (a *AsyncCtrl) IsClosed() bool {
	return a.closed
}

func (a *AsyncCtrl) Context() context.Context {
	return a.ctx
}

//The LState is not goroutine-safe. It is recommended to use one LState per goroutine and communicate between goroutines by using channels.
func (a *AsyncCtrl) LockLState() {
	a.mu.Lock()
}

func (a *AsyncCtrl) UnlockLState() {
	a.mu.Unlock()
}
func (a *AsyncCtrl) SafeCall(L *lua.LState, cp lua.P, args ...lua.LValue) (err error) {
	if a.IsClosed() {
		return nil
	}
	a.LockLState()
	defer a.UnlockLState()
	return L.CallByParam(cp, args...)
}

func (a *AsyncCtrl) NewGoRoutine(fn func()) {
	if a.IsClosed() {
		return
	}
	a.Add(1)
	go func() {
		defer a.Done()
		fn()
	}()
}
