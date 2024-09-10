package subject

// Observer is the interface that wraps the Notify method.
type Observer[T any] interface {
	// Notify notifies the observer of a change.
	//
	// Parameters:
	//   - change: The change that occurred.
	//
	// Returns:
	// 	- bool: True if the receiver is not nil, false otherwise.
	Notify(change T) bool
}

// ReactiveObserver is a type that acts as a simple observer that calls a function
// when a change occurs.
type ReactiveObserver[T any] struct {
	// event is the event to call when a change occurs.
	event func(T)
}

// Notify implements the Observer interface.
func (r *ReactiveObserver[T]) Notify(change T) bool {
	if r == nil {
		return false
	}

	r.event(change)

	return true
}

// NewReactiveObserver creates a new ReactiveObserver.
//
// Parameters:
//   - event: The event to call when a change occurs.
//
// Returns:
//   - *ReactiveObserver[T]: A new ReactiveObserver. Never returns nil.
func NewReactiveObserver[T any](event func(T)) *ReactiveObserver[T] {
	return &ReactiveObserver[T]{
		event: event,
	}
}
