package sync_wrapper

import "sync"

type SyncKVMap[K, T any] struct {
	sync.Map
}

func NewSyncKVMap[K, T any]() *SyncKVMap[K, T] {
	return &SyncKVMap[K, T]{sync.Map{}}
}

func (m *SyncKVMap[K, T]) Set(key K, value T) {
	m.Map.Store(key, value)
}

func (m *SyncKVMap[K, T]) Get(key K) (val T, ok bool) {
	v, ok := m.Map.Load(key)
	if ok {
		return v.(T), true
	} else {
		return val, false
	}
}

func (m *SyncKVMap[K, T]) UnsafeGetAndUpdate(key K, operator func(val T) T) {
	var oldVal any
	var empty T
	var ok bool
	oldVal, ok = m.Map.Load(key)
	if ok {
		newValue := operator(oldVal.(T))
		m.Map.Store(key, newValue)
	} else {
		newValue := operator(empty)
		m.Map.Store(key, newValue)
	}
}

func (m *SyncKVMap[K, T]) Delete(key K) {
	m.Map.Delete(key)
}

func (m *SyncKVMap[K, T]) GetOrSet(key K, value T) (val T, ok bool) {
	v, ok := m.Map.LoadOrStore(key, value)
	return v.(T), ok
}

func (m *SyncKVMap[K, T]) GetAndDelete(key K) (val T, ok bool) {
	v, ok := m.Map.LoadAndDelete(key)
	if ok {
		return v.(T), true
	} else {
		return val, false
	}
}

func (m *SyncKVMap[K, T]) Iter(fn func(k K, v T) (continueInter bool)) {
	m.Map.Range(func(key, value any) bool {
		return fn(key.(K), value.(T))
	})
}
