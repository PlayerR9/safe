package rw_safe

import (
	"sync"
)

// Safe is a rw mutex protected variable.
type Safe[T any] struct {
	// value is the value of the safe variable.
	value T

	// mu is the mutex to synchronize access to the safe variable.
	mu sync.RWMutex
}

// Copy is a method that returns a copy of the safe variable.
//
// Returns:
//   - *Safe[T]: A copy of the safe variable. Never returns nil.
//
// If receiver is nil, then a new variable will be created instead initialized
// with the value as its zero value.
func (s *Safe[T]) Copy() *Safe[T] {
	if s == nil {
		return &Safe[T]{
			value: *new(T),
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return &Safe[T]{
		value: s.value,
	}
}

// NewSafe creates a new safe variable.
//
// Parameters:
//   - value: The value of the safe variable.
//
// Returns:
//   - *Safe[T]: A new safe variable. Never returns nil.
func NewSafe[T any](value T) *Safe[T] {
	s := &Safe[T]{
		value: value,
	}

	return s
}

// Set sets the value of the safe variable.
//
// Parameters:
//   - value: The value to set the safe variable to.
//
// Returns:
//   - bool: True if the receiver is not nil. False otherwise.
func (s *Safe[T]) Set(value T) bool {
	if s == nil {
		return false
	}

	s.mu.Lock()
	s.value = value
	s.mu.Unlock()

	return true
}

// Get gets the value of the safe variable.
//
// Returns:
//   - T: The value of the safe variable.
//
// If the receiver is nil, then the zero value is returned instead.
func (s *Safe[T]) Get() T {
	if s == nil {
		return *new(T)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.value
}

// Modifyvalue modifies the value of the safe variable.
//
// Parameters:
//   - f: The function to modify the value of the safe variable.
//
// If 'f' or the receiver are nil, then nothing is done.
func (s *Safe[T]) Modifyvalue(f func(T) T) {
	if s == nil || f == nil {
		return
	}

	s.mu.RLock()
	curr := s.value
	s.mu.RUnlock()

	new := f(curr)

	s.mu.Lock()
	s.value = new
	s.mu.Unlock()
}

// DoRead is a method of the safe variable type. It is used to perform a read
// operation on the value stored in the safe variable.
// Through the function parameter, the caller can access the value in a
// read-only manner.
//
// Parameters:
//   - f: A function that takes a value of type T as a parameter and
//     returns nothing.
//
// If 'f' or receiver are nil, then nothing is done.
func (s *Safe[T]) DoRead(f func(T)) {
	if s == nil || f == nil {
		return
	}

	s.mu.RLock()
	value := s.value
	s.mu.RUnlock()

	f(value)
}
