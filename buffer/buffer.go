package buffer

import (
	"sync"

	rws "github.com/PlayerR9/safe/rw_safe"
)

// BufferCondition is an enumeration of the possible conditions of the Buffer.
type BufferCondition int

const (
	// IsEmpty indicates that the Buffer is empty.
	IsEmpty BufferCondition = iota

	// IsRunning indicates that the Buffer is running.
	IsRunning
)

// String implements Common.Enum interface.
func (bc BufferCondition) String() string {
	return [...]string{"IsEmpty", "IsRunning"}[bc]
}

// Buffer is a thread-safe, generic data structure that allows multiple
// goroutines to produce and consume elements in a synchronized manner.
// It is implemented as a queue and uses channels to synchronize the
// goroutines.
// The Buffer should be initialized with the Init method before use.
type Buffer[T any] struct {
	// q is a pointer to the SafeQueue that stores the elements of the Buffer.
	q *SafeQueue[T]

	// sendTo is a channel that receives messages and sends them to the Buffer.
	sendTo chan T

	// receiveFrom is a channel that receives messages from the Buffer and
	// sends them to the consumer.
	receiveFrom chan T

	// wg is a WaitGroup that is used to wait for the goroutines to finish.
	wg sync.WaitGroup

	// locker is a pointer to the RWSafe that synchronizes the Buffer.
	locker *rws.Locker[BufferCondition]
}

// Start implements the Runner interface.
func (b *Buffer[T]) Start() {
	if b.locker != nil {
		return
	}

	b.locker = rws.NewLocker[BufferCondition]()
	b.locker.SetSubject(IsEmpty, true, true)
	b.locker.SetSubject(IsRunning, true, true)

	b.q = NewSafeQueue[T]()
	b.q.ObserveSize(func(val int) {
		ok := b.locker.ChangeValue(IsEmpty, val == 0)
		if !ok {
			panic("unable to change value")
		}
	})

	b.sendTo = make(chan T)
	b.receiveFrom = make(chan T)

	b.wg.Add(2)

	go b.listenForIncomingMessages()
	go b.sendMessagesFromBuffer()
}

// Close implements the Runner interface.
func (b *Buffer[T]) Close() {
	if b.sendTo == nil {
		return
	}

	close(b.sendTo)
	b.sendTo = nil

	b.wg.Wait()

	close(b.receiveFrom)
	b.receiveFrom = nil
}

// IsClosed implements the Runner interface.
func (b *Buffer[T]) IsClosed() bool {
	return b.locker == nil
}

// Send implements the Sender interface.
func (b *Buffer[T]) Send(msg T) bool {
	if b.sendTo == nil {
		return false
	}

	b.sendTo <- msg

	return true
}

// Receive implements the Receiver interface.
func (b *Buffer[T]) Receive() (T, bool) {
	if b.receiveFrom == nil {
		return *new(T), false
	}

	msg, ok := <-b.receiveFrom
	if !ok {
		return *new(T), false
	}

	return msg, true
}

// NewBuffer creates a new Buffer instance.
//
// Parameters:
//   - bufferSize: The size of the buffer for the send and receive channels.
//     Must be a non-negative integer. If a negative integer is provided,
//     the method will panic with an *ers.InvalidParameterError.
//
// Returns:
//   - *Buffer: A pointer to the newly created Buffer instance.
//   - error: An error of type *ers.InvalidParameterError if
//     the buffer size is negative.
//
// Information: To close the buffer, just close the send-only channel.
// Once that is done, a cascade of events will happen:
//   - The goroutine that listens for incoming messages will stop listening
//     and exit.
//   - The goroutine that sends messages from the Buffer to the receive
//     channel will stop sending messages once the Buffer is empty, and then exit.
//   - The Buffer will be cleaned up.
//
// Of course, a Close method is also provided to manually close the Buffer but
// it is not necessary to call it if the send-only channel is closed.
func NewBuffer[T any]() *Buffer[T] {
	return &Buffer[T]{}
}

// listenForIncomingMessages is a method of the Buffer type that listens for
// incoming messages from the receiveChannel and enqueues them in the Buffer.
//
// It must be run in a separate goroutine to avoid blocking the main thread.
func (b *Buffer[T]) listenForIncomingMessages() {
	defer b.wg.Done()

	for msg := range b.sendTo {
		b.q.Enqueue(msg)
	}

	b.locker.ChangeValue(IsRunning, false)
}

// sendMessagesFromBuffer is a method of the Buffer type that sends
// messages from the Buffer to the sendChannel.
//
// It must be run in a separate goroutine to avoid blocking the main thread.
func (b *Buffer[T]) sendMessagesFromBuffer() {
	defer b.wg.Done()

	for {
		value, ok := b.locker.Get(IsRunning)
		if !ok {
			panic("unable to get whether the buffer is running or not")
		}

		if !value {
			break
		}

		ok = b.locker.DoFunc(func(m map[BufferCondition]bool) bool {
			for {
				isEmpty, ok := b.sendSingleMessage()
				if !ok || isEmpty {
					break
				}
			}

			return m[IsRunning]
		})
		if !ok {
			break
		}
	}

	for {
		isEmpty, _ := b.sendSingleMessage()
		if isEmpty {
			break
		}
	}

	b.locker = nil
	b.q = nil
}

// sendSingleMessage is a method of the Buffer type that sends a single message
// from the Buffer to the send channel.
//
// Returns:
//   - bool: A boolean indicating if the queue is empty.
//   - bool: A boolean indicating if a message was sent successfully.
func (b *Buffer[T]) sendSingleMessage() (bool, bool) {
	msg, ok := b.q.Peek()
	if !ok {
		return true, true
	}

	select {
	case b.receiveFrom <- msg:
		_, ok := b.q.Dequeue()
		if !ok {
			return true, false
		}

		return false, true
	default:
		return false, false
	}
}

// CleanBuffer removes all elements from the Buffer, effectively resetting
// it to an empty state. Precalculated elements are kept as they are no longer
// in the buffer but in the channel. It locks the firstMutex to ensure
// thread-safety during the operation.
//
// This method is safe for concurrent use by multiple goroutines.
func (b *Buffer[T]) CleanBuffer() {
	if b.q == nil {
		return
	}

	b.q.Clear()
}
