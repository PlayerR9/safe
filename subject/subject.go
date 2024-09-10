package subject

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

// NewSubject creates a new subject.
//
// Parameters:
//   - state: The initial state of the subject.
//
// Returns:
//   - *Subject[T]: A new subject. Never returns nil.
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
//
// Returns:
//   - bool: True if receiver is not nil, false otherwise.
func (s *Subject[T]) Attach(o Observer[T]) bool {
	if s == nil {
		return false
	}

	if o == nil {
		return true
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.observers = append(s.observers, o)

	return true
}

// Set sets the state of the subject.
//
// Parameters:
//   - state: The new state of the subject.
//
// Returns:
//   - bool: True if the receiver is not nil and all observers were notified,
//     false otherwise.
func (s *Subject[T]) Set(state T) bool {
	if s == nil {
		return false
	}

	s.mu.Lock()
	s.state = state
	s.mu.Unlock()

	n := s.NotifyAll()
	return n == 0
}

// State gets the state of the subject.
//
// Returns:
//   - T: The state of the subject.
//
// If recever is nil, then the zero value is returned.
func (s *Subject[T]) State() T {
	if s == nil {
		return *new(T)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.state
}

// ModifyState modifies the state of the subject.
//
// Parameters:
//   - f: The function to modify the state of the subject.
//
// Returns:
//   - bool: True if 'f' is nil or all observers were successfully notified,
//     false otherwise.
//
// If recever is nil, false is returned.
func (s *Subject[T]) ModifyState(f func(T) T) bool {
	if s == nil {
		return false
	} else if f == nil {
		return true
	}

	s.mu.RLock()
	curr := s.state
	s.mu.RUnlock()

	new := f(curr)

	s.mu.Lock()
	s.state = new
	s.mu.Unlock()

	n := s.NotifyAll()
	return n == 0
}

// NotifyAll notifies all observers of a change.
//
// Returns:
//   - int: The number of observers that were not notified.
//
// If this function returns a non-zero value, then an observer had a nil receiver.
func (s *Subject[T]) NotifyAll() int {
	s.mu.RLock()
	state := s.state
	observerSize := len(s.observers)
	s.mu.RUnlock()

	if observerSize == 0 {
		return 0
	}

	var count int
	var count_mu sync.RWMutex

	var wg sync.WaitGroup

	wg.Add(len(s.observers))

	for _, observer := range s.observers {
		fn := func(observer Observer[T]) {
			defer wg.Done()

			ok := observer.Notify(state)
			if ok {
				return
			}

			count_mu.Lock()
			defer count_mu.Unlock()

			count++
		}

		go fn(observer)
	}

	wg.Wait()

	count_mu.Lock()
	defer count_mu.Unlock()

	return count
}

// DoRead is a method of the Subject type. It is used to perform a read
// operation on the value stored in the Subject.
// Through the function parameter, the caller can access the value in a
// read-only manner.
//
// Parameters:
//   - f: A function that takes a value of type T as a parameter and
//     returns nothing.
//
// If 'f' or receiver are nil, then nothing is done.
func (s *Subject[T]) DoRead(f func(T)) {
	if s == nil || f == nil {
		return
	}

	s.mu.RLock()
	value := s.state
	s.mu.RUnlock()

	f(value)
}

// SetObserver sets the observer of the subject.
//
// Parameters:
//   - action: The action to set as the observer.
//
// if 'action' or the receiver are nil, then nothing is done.
func (s *Subject[T]) SetObserver(action func(T)) {
	if s == nil || action == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.observers = append(s.observers, NewReactiveObserver(action))
}

// Copy creates a shallow copy of the subject.
//
// Returns:
//   - *Subject[T]: A shallow copy of the subject. Never returns nil.
//
// It is important to note that the observers are not copied and so,
// they have to be reattached to the new subject.
//
// If the receiver is nil, a new subject is returned that has its value
// initialized with its zero value and with no observers.
func (s *Subject[T]) Copy() *Subject[T] {
	if s == nil {
		return &Subject[T]{
			observers: make([]Observer[T], 0),
			state:     *new(T),
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return &Subject[T]{
		observers: make([]Observer[T], 0),
		state:     s.state,
	}
}
