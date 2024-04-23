package generics

import (
	"fmt"
	"sync"
)

// Map is like a Go map[interface{}]interface{} but is safe for concurrent use
// by multiple goroutines without additional locking or coordination.
// Loads, stores, and deletes run in amortized constant time.
//
// The Map type is specialized. Most code should use a plain Go map instead,
// with separate locking or coordination, for better type safety and to make it
// easier to maintain other invariants along with the map content.
//
// The Map type is optimized for two common use cases: (1) when the entry for a given
// key is only ever written once but read many times, as in caches that only grow,
// or (2) when multiple goroutines read, write, and overwrite entries for disjoint
// sets of keys. In these two cases, use of a Map may significantly reduce lock
// contention compared to a Go map paired with a separate Mutex or RWMutex.
//
// The zero Map is empty and ready for use. A Map must not be copied after first use.
//
// In the terminology of the Go memory model, Map arranges that a write operation
// “synchronizes before” any read operation that observes the effect of the write, where
// read and write operations are defined as follows.
// Load, LoadAndDelete, LoadOrStore, Swap, CompareAndSwap, and CompareAndDelete
// are read operations; Delete, LoadAndDelete, Store, and Swap are write operations;
// LoadOrStore is a write operation when it returns loaded set to false;
// CompareAndSwap is a write operation when it returns swapped set to true;
// and CompareAndDelete is a write operation when it returns deleted set to true.
//
// Specially, generics are supported here for "sync.Map".
type SyncMap[K comparable, V any] struct {
	mapping sync.Map
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
// The old value must be of a comparable type.
//
// If there is no current value for key in the map, CompareAndDelete
// returns false (even if the old value is the nil interface value).
func (m *SyncMap[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.mapping.CompareAndDelete(key, old)
}

// CompareAndSwap swaps the old and new values for key
// if the value stored in the map is equal to old.
// The old value must be of a comparable type.
func (m *SyncMap[K, V]) CompareAndSwap(key, old K, new V) bool {
	return m.mapping.CompareAndSwap(key, old, new)
}

// Delete deletes the value for a key.
func (m *SyncMap[K, V]) Delete(key K) {
	m.mapping.Delete(key)
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	val, ok := m.mapping.Load(key)
	if !ok {
		return
	}
	value, success := val.(V)
	if val != nil && !success {
		panic(fmt.Sprintf("Load: Failed to convert val into type %T; val = %#v", value, val))
	}
	return
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *SyncMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	val, loaded := m.mapping.LoadAndDelete(key)
	if !loaded {
		return
	}
	value, success := val.(V)
	if val != nil && !success {
		panic(fmt.Sprintf("LoadAndDelete: Failed to convert val into type %T; val = %#v", value, val))
	}
	return
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *SyncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	act, loaded := m.mapping.LoadOrStore(key, value)
	if !loaded {
		return
	}
	actual, success := act.(V)
	if act != nil && !success {
		panic(fmt.Sprintf("LoadOrStore: Failed to convert act into type %T; act = %#v", actual, act))
	}
	return
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the Map's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently (including by f), Range may reflect any
// mapping for that key from any point during the Range call. Range does not
// block other methods on the receiver; even f itself may call any method on m.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	m.mapping.Range(func(key, value any) bool {
		k, success := key.(K)
		if key != nil && !success {
			panic(fmt.Sprintf("Range: Failed to convert key into type %T; key = %#v", k, key))
		}
		v, success := value.(V)
		if value != nil && !success {
			panic(fmt.Sprintf("Range: Failed to convert value into type %T; value = %#v", v, value))
		}
		return f(k, v)
	})
}

// Store sets the value for a key.
func (m *SyncMap[K, V]) Store(key K, value V) {
	m.mapping.Store(key, value)
}

// Swap swaps the value for a key and returns the previous value if any.
// The loaded result reports whether the key was present.
func (m *SyncMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	old, loaded := m.mapping.Swap(key, value)
	if !loaded {
		return
	}
	previous, success := old.(V)
	if old != nil && !success {
		panic(fmt.Sprintf("Swap: Failed to convert old into type %T; old = %#v", previous, old))
	}
	return
}
