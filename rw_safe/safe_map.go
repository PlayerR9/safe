package rw_safe

import (
	"iter"
	"sync"
)

// SafeMap is a thread-safe map.
type SafeMap[T comparable, U any] struct {
	// m is the underlying map.
	m map[T]U

	// mu is the mutex to synchronize map access.
	mu sync.RWMutex
}

// Copy is a method that returns a copy of the SafeMap.
//
// Returns:
//   - *SafeMap[T, U]: A copy of the SafeMap.
func (sm *SafeMap[T, U]) Copy() *SafeMap[T, U] {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	newMap := make(map[T]U, len(sm.m))
	for key, value := range sm.m {
		newMap[key] = value
	}

	return &SafeMap[T, U]{
		m: newMap,
	}
}

// Entry is a method that returns an iterator over the entries in the SafeMap.
//
// Returns:
//   - iter.Seq2[T, U]: An iterator over the entries in the SafeMap.
func (sm *SafeMap[T, U]) Entry() iter.Seq2[T, U] {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	fn := func(yield func(T, U) bool) {
		for key, value := range sm.m {
			if !yield(key, value) {
				return
			}
		}
	}

	return fn
}

// NewSafeMap creates a new SafeMap.
//
// Returns:
//   - *SafeMap[T, U]: A new SafeMap.
func NewSafeMap[T comparable, U any]() *SafeMap[T, U] {
	return &SafeMap[T, U]{
		m: make(map[T]U),
	}
}

// Get retrieves a value from the map.
//
// Parameters:
//   - key: The key to retrieve the value.
//
// Returns:
//   - U: The value associated with the key.
//   - bool: A boolean indicating if the key exists in the map.
func (sm *SafeMap[T, U]) Get(key T) (U, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	val, ok := sm.m[key]
	return val, ok
}

// Set sets a value in the map.
//
// Parameters:
//   - key: The key to set the value.
//   - val: The value to set.
func (sm *SafeMap[T, U]) Set(key T, val U) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.m[key] = val
}

// Delete removes a key from the map.
//
// Parameters:
//   - key: The key to remove.
func (sm *SafeMap[T, U]) Delete(key T) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.m, key)
}

// Len returns the number of elements in the map.
//
// Returns:
//   - int: The number of elements in the map.
func (sm *SafeMap[T, U]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.m)
}

// Clear removes all elements from the map.
func (sm *SafeMap[T, U]) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.m = make(map[T]U)
}

// ScanFunc is a function that can be applied to all elements in the map.
//
// Parameters:
//   - key: The key of the element.
//   - value: The value of the element.
//
// Returns:
//   - bool: A boolean indicating if the scan should continue.
//   - error: An error if the scan should stop.
type ScanFunc[T, U any] func(key T, value U) (bool, error)

// Scan applies a read-only function to all elements in the map.
//
// Parameters:
//   - f: The function to apply to all elements.
//
// Returns:
//   - bool: A boolean indicating if the scan completed successfully.
//   - error: An error if the scan failed.
func (sm *SafeMap[T, U]) Scan(f ScanFunc[T, U]) (bool, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for key, value := range sm.m {
		ok, err := f(key, value)
		if err != nil {
			return false, err
		}

		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// GetMap returns the underlying map.
//
// Returns:
//   - map[T]U: The underlying map.
func (sm *SafeMap[T, U]) GetMap() map[T]U {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	mapCopy := make(map[T]U, len(sm.m))
	for key, value := range sm.m {
		mapCopy[key] = value
	}

	return mapCopy
}
