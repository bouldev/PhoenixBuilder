package sync_wrapper

import (
	"sync"

	"github.com/google/uuid"
)

type detachableCallbackHandle[T any] struct {
	isDetached bool
	callback   func(T)
}

type HybridListener[T any] struct {
	mu                  sync.RWMutex
	fixedCallbacks      []func(T)
	detachableCallbacks map[string]*detachableCallbackHandle[T]
}

func (l *HybridListener[T]) Len() int {
	return len(l.fixedCallbacks) + len(l.detachableCallbacks)
}

func (l *HybridListener[T]) Call(data T) {
	if l.Len() == 0 {
		return
	}
	for _, cb := range l.fixedCallbacks {
		cb(data)
	}
	if len(l.detachableCallbacks) == 0 {
		return
	}
	l.mu.RLock()
	cpyListener := make([]*detachableCallbackHandle[T], 0, len(l.detachableCallbacks))
	for _, v := range l.detachableCallbacks {
		cpyListener = append(cpyListener, v)
	}
	l.mu.RUnlock()
	for _, l := range cpyListener {
		if l.isDetached {
			continue
		}
		l.callback(data)
	}
}

func (l *HybridListener[T]) SetNonBlockingFixListener(cb func(T)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fixedCallbacks = append(l.fixedCallbacks, cb)
}

func (l *HybridListener[T]) AttachNonBlockingDetachableListener(cb func(T)) (detach func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	uid := uuid.NewString()
	h := &detachableCallbackHandle[T]{
		isDetached: false,
		callback:   cb,
	}
	l.detachableCallbacks[uid] = h
	detach = func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if h.isDetached {
			return
		}
		delete(l.detachableCallbacks, uid)
	}
	return detach
}

func NewHybridListener[T any]() *HybridListener[T] {
	return &HybridListener[T]{
		mu:                  sync.RWMutex{},
		fixedCallbacks:      make([]func(T), 0),
		detachableCallbacks: make(map[string]*detachableCallbackHandle[T]),
	}
}
