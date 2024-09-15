package rw_safe

import (
	"sync"
)

// Subject is the subject that observers observe.
type Subject[T any] struct {
	// observers is the list of observers.
	observers []Observer[T]

	// state is the state of the subject.
	state T

	// mu is the mutex to synchronize access to the subject.
	mu sync.RWMutex
}

// Copy is a method that returns a copy of the subject.
//
// Returns:
//   - *Subject[T]: A copy of the subject.
//
// However, the obsevers are not copied.
func (s *Subject[T]) Copy() *Subject[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &Subject[T]{
		observers: make([]Observer[T], 0),
		state:     s.state,
	}
}

// NewSubject creates a new subject.
//
// Parameters:
//   - state: The initial state of the subject.
//
// Returns:
//
//	*Subject[T]: A new subject.
func NewSubject[T any](state T) *Subject[T] {
	return &Subject[T]{
		observers: make([]Observer[T], 0),
		state:     state,
	}
}

// Attach attaches an observer to the subject.
//
// Parameters:
//   - o: The observer to attach.
func (s *Subject[T]) Attach(o Observer[T]) {
	if o == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.observers = append(s.observers, o)
}

// Set sets the state of the subject.
//
// Parameters:
//   - state: The new state of the subject.
func (s *Subject[T]) Set(state T) {
	s.mu.Lock()
	s.state = state
	s.mu.Unlock()

	s.NotifyAll()
}

// Get gets the state of the subject.
//
// Returns:
//   - T: The state of the subject.
func (s *Subject[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.state
}

// ModifyState modifies the state of the subject.
//
// Parameters:
//   - f: The function to modify the state of the subject.
func (s *Subject[T]) ModifyState(f func(T) T) {
	s.mu.RLock()
	curr := s.state
	s.mu.RUnlock()

	new := f(curr)

	s.mu.Lock()
	s.state = new
	s.mu.Unlock()

	s.NotifyAll()
}

// NotifyAll notifies all observers of a change.
func (s *Subject[T]) NotifyAll() {
	s.mu.RLock()
	state := s.state
	observerSize := len(s.observers)
	s.mu.RUnlock()

	if observerSize == 0 {
		return
	}

	var wg sync.WaitGroup

	wg.Add(len(s.observers))

	for _, observer := range s.observers {
		go func(observer Observer[T]) {
			defer wg.Done()

			observer.Notify(state)
		}(observer)
	}

	wg.Wait()
}

// DoRead is a method of the Subject type. It is used to perform a read
// operation on the value stored in the Subject.
// Through the function parameter, the caller can access the value in a
// read-only manner.
//
// Parameters:
//
//   - f: A function that takes a value of type T as a parameter and
//     returns nothing.
func (s *Subject[T]) DoRead(f func(T)) {
	s.mu.RLock()
	value := s.state
	s.mu.RUnlock()

	f(value)
}

// SetObserver sets the observer of the subject.
//
// Parameters:
//   - action: The action to set as the observer.
func (s *Subject[T]) SetObserver(action func(T)) {
	if action == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.observers = append(s.observers, NewReactiveObserver(action))
}
