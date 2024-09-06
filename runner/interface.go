package runner

// Runner is an interface that defines the behavior of a type that can be started,
// stopped, and waited for.
type Runner interface {
	// Start starts the runner. If it is already running, nothing happens.
	Start()

	// Close closes the runner. If it is not running, nothing happens.
	Close()

	// IsClosed returns true if the runner is closed, false otherwise.
	IsClosed() bool
}

// Sender is the interface that wraps the Send method.
type Sender[T any] interface {
	// Send sends a message to the Buffer.
	//
	// Parameters:
	//   - msg: The message to send.
	//
	// Returns:
	//   - bool: False if the Buffer is closed, true otherwise.
	Send(msg T) bool
}

// SenderRunner is the interface that wraps the Send method and the Runner interface.
type SenderRunner[T any] interface {
	Sender[T]
	Runner
}
