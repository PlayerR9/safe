package buffer

// Receiver is the interface that wraps the Receive method.
type Receiver[T any] interface {
	// Receive receives a message from the Buffer.
	//
	// Returns:
	//   - T: The message received.
	//   - bool: False if the Buffer is closed, true otherwise.
	Receive() (T, bool)
}
