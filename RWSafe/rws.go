package RWSafe

import (
	"sync"

	uc "github.com/PlayerR9/MyGoLib/Units/common"
)

// Safe is a rw mutex protected variable.
type Safe[T any] struct {
	// value is the value of the safe variable.
	value T

	// mu is the mutex to synchronize access to the safe variable.
	mu sync.RWMutex
}

// Copy implements the Copier interface.
func (s *Safe[T]) Copy() uc.Copier {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sCopy := &Safe[T]{
		value: s.value,
	}

	return sCopy
}

// NewSafe creates a new safe variable.
//
// Parameters:
//   - value: The value of the safe variable.
//
// Returns:
//   - *Safe[T]: A new safe variable.
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
func (s *Safe[T]) Set(value T) {
	s.mu.Lock()
	s.value = value
	s.mu.Unlock()
}

// Get gets the value of the safe variable.
//
// Returns:
//   - T: The value of the safe variable.
func (s *Safe[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.value
}

// Modifyvalue modifies the value of the safe variable.
//
// Parameters:
//   - f: The function to modify the value of the safe variable.
func (s *Safe[T]) Modifyvalue(f func(T) T) {
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
//
//   - f: A function that takes a value of type T as a parameter and
//     returns nothing.
func (s *Safe[T]) DoRead(f func(T)) {
	s.mu.RLock()
	value := s.value
	s.mu.RUnlock()

	f(value)
}
