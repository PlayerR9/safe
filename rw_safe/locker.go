package rw_safe

import (
	"fmt"
	"sync"
)

// Conditioner is an interface that represents a condition.
type Conditioner interface {
	~int

	fmt.Stringer
}

// Locker is a thread-Subject locker that allows multiple goroutines to wait for a condition.
type Locker[T Conditioner] struct {
	// cond is the condition variable.
	cond *sync.Cond

	// elems is the list of elements.
	subjects map[T]*Subject[bool]

	// mu is the mutex to synchronize map access.
	mu sync.RWMutex
}

// NewLocker creates a new Locker.
//
// Use Locker.Set for observer boolean predicates.
//
// Parameters:
//   - keys: The keys to initialize the locker.
//
// Returns:
//   - *Locker[T]: A new Locker.
//
// Behaviors:
//   - All the predicates are initialized to true.
func NewLocker[T Conditioner]() *Locker[T] {
	l := &Locker[T]{
		cond:     sync.NewCond(&sync.Mutex{}),
		subjects: make(map[T]*Subject[bool]),
	}

	return l
}

// SetSubject adds a new subject to the locker.
//
// Parameters:
//   - key: The key to add.
//   - subject: The subject to add.
//   - broadcast: A flag indicating whether the subject should broadcast or signal.
//
// Behaviors:
//   - If the subject is nil, it will not be added.
//   - It overwrites the existing subject if the key already exists.
func (l *Locker[T]) SetSubject(key T, value bool, broadcast bool) {
	subject := NewSubject(value)

	if broadcast {
		subject.SetObserver(func(b bool) {
			l.cond.L.Lock()
			defer l.cond.L.Unlock()

			l.cond.Broadcast()
		})
	} else {
		subject.SetObserver(func(b bool) {
			l.cond.L.Lock()
			defer l.cond.L.Unlock()

			l.cond.Signal()
		})
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.subjects[key] = subject
}

// ChangeValue changes the value of a subject.
//
// Parameters:
//   - key: The key to change the value.
//   - value: The new value.
//
// Returns:
//   - bool: True if the key exists, false otherwise.
func (l *Locker[T]) ChangeValue(key T, value bool) bool {
	l.mu.Lock()
	subject, ok := l.subjects[key]
	l.mu.Unlock()

	if !ok {
		return false
	} else {
		subject.Set(value)
		return true
	}

}

// hasFalse is a private method that checks if at least one of the conditions is false.
//
// Returns:
//   - map[T]bool: A copy of the map of conditions.
//   - bool: True if at least one of the conditions is false, false otherwise.
func (l *Locker[T]) hasFalse() (map[T]bool, bool) {
	l.mu.RLock()

	mapCopy := make(map[T]bool)

	for key, value := range l.subjects {
		mapCopy[key] = value.Get()
	}
	l.mu.RUnlock()

	for _, value := range mapCopy {
		if !value {
			return mapCopy, true
		}
	}

	return mapCopy, false
}

// Get returns the value of a predicate.
//
// Parameters:
//   - key: The key to get the value.
//
// Returns:
//   - bool: The value of the predicate.
//   - bool: True if the key exists, false otherwise.
func (l *Locker[T]) Get(key T) (bool, bool) {
	l.mu.RLock()
	val, ok := l.subjects[key]
	l.mu.RUnlock()

	if ok {
		return val.Get(), true
	} else {
		return false, false
	}
}

// DoFunc is a function that executes a function while waiting for the condition to be false.
//
// Parameters:
//   - f: The function to execute that takes a map of conditions as a parameter and returns
//     true if the function should exit, false otherwise.
//
// Returns:
//   - bool: True if the function should exit, false otherwise.
func (l *Locker[T]) DoFunc(f func(map[T]bool) bool) bool {
	l.cond.L.Lock()

	var mapCopy map[T]bool
	var ok bool

	for {
		mapCopy, ok = l.hasFalse()
		if ok {
			l.cond.L.Unlock()
			break
		}

		l.cond.Wait()
	}

	shouldContinue := f(mapCopy)

	return shouldContinue
}
