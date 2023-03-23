package sync_wrapper

import "sync"

type SyncMap[T any] struct {
	sync.Map
}

func NewInstanceMap[T any]() *SyncMap[T] {
	return &SyncMap[T]{sync.Map{}}
}

func (m *SyncMap[T]) Set(key string, value T) {
	m.Map.Store(key, value)
}

func (m *SyncMap[T]) Get(key string) (val T, ok bool) {
	v, ok := m.Map.Load(key)
	if ok {
		return v.(T), true
	} else {
		return val, false
	}
}

func (m *SyncMap[T]) Delete(key string) {
	m.Map.Delete(key)
}

func (m *SyncMap[T]) Iter(fn func(k string, v T) (continueInter bool)) {
	m.Map.Range(func(key, value any) bool {
		return fn(key.(string), value.(T))
	})
}
