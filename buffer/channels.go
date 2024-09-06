package buffer

import (
	"sync"

	gcslc "github.com/PlayerR9/go-commons/slices"
	ur "github.com/PlayerR9/safe/runner"
	rws "github.com/PlayerR9/safe/rw_safe"
)

// DiscardAnyMessage is a function that discards all messages from a receiver.
//
// Parameters:
//   - receiver: The receiver of messages.
//
// Behaviors:
//   - Use go DiscardAnyMessage(receiver) to discard all messages from the receiver.
func DiscardAnyMessage[T any](receiver Receiver[T]) {
	if receiver == nil {
		return
	}

	for {
		_, ok := receiver.Receive()
		if !ok {
			break
		}
	}
}

// Redirect is a handler that redirects messages from a receiver to multiple senders.
type Redirect[T any] struct {
	// receiver is the receiver of messages.
	receiver Receiver[T]

	// senders is a slice of senders of messages.
	senders []ur.SenderRunner[T]

	// isClosed is a flag that indicates if the handler is closed.
	isClosed *rws.Safe[bool]
}

// NewRedirect creates a new redirect handler.
//
// Parameters:
//   - receiver: The receiver of messages.
//   - senders: The senders of messages.
//
// Returns:
//   - *Redirect: The new redirect handler.
//
// Behaviors:
//   - It ignores nil senders.
//   - Because it closes automatically, there is no Close() method.
//     Thus, if the receiver is closed, the handler will close all senders in a
//     cascading manner.
//   - If no senders are provided, the handler will discard all messages from the receiver.
func NewRedirect[T any](receiver Receiver[T], senders ...ur.SenderRunner[T]) *Redirect[T] {
	senders = gcslc.SliceFilter(senders, func(sender ur.SenderRunner[T]) bool {
		return sender != nil
	})

	return &Redirect[T]{
		receiver: receiver,
		senders:  senders,
	}
}

// run is a private method of Redirect that redirects messages from the receiver to the senders.
//
// Parameters:
//   - senderCopy: A copy of the senders.
func (r *Redirect[T]) run() {
	defer r.isClosed.Set(true)

	for {
		msg, ok := r.receiver.Receive()
		if !ok {
			// receiver is closed.
			break
		}

		for _, sender := range r.senders {
			go sender.Send(msg)
		}
	}

	var wg sync.WaitGroup

	wg.Add(len(r.senders))

	for _, sender := range r.senders {
		go sender.Close()
	}

	wg.Wait()
}

// Run is a method that runs the handler.
func (r *Redirect[T]) Run() {
	if r.receiver == nil || r.IsRunning() {
		return
	}

	r.isClosed = rws.NewSafe(false)

	if len(r.senders) == 0 {
		go func() {
			defer r.isClosed.Set(true)

			DiscardAnyMessage(r.receiver)
		}()
	} else {
		var wg sync.WaitGroup

		wg.Add(len(r.senders))

		for _, sender := range r.senders {
			go func(sender ur.SenderRunner[T]) {
				defer wg.Done()

				sender.Start()
			}(sender)
		}

		wg.Wait()

		go r.run()
	}
}

// IsRunning is a method that returns true if the handler is running.
//
// Returns:
//   - bool: True if the handler is running, false otherwise.
func (r *Redirect[T]) IsRunning() bool {
	return r.isClosed != nil && !r.isClosed.Get()
}

// ChannelThrough is a type of buffer that sends messages from multiple receivers to a
// single sender.
type ChannelThrough[T any] struct {
	// receivers is a slice of receivers of messages.
	receivers []Receiver[T]

	// sender is the sender of messages.
	sender ur.SenderRunner[T]

	// buffer is the buffer that stores the messages.
	buffer *Buffer[T]

	// isClosed is a flag that indicates if the handler is closed.
	isClosed *rws.Safe[bool]
}

// NewChannelThrough creates a new channel through buffer.
//
// Parameters:
//   - sender: The sender of messages.
//   - receivers: The receivers of messages.
//
// Returns:
//   - *ChannelThrough: The new channel through buffer.
//
// Behaviors:
//   - It ignores nil receivers.
//   - If the sender is nil, it will discard all messages from the receivers.
//   - If no receivers are provided, the buffer will close immediately.
//   - Because it closes automatically, there is no Close() method.
//     Thus, if all receivers are closed, the handler will close the sender in a
//     cascading manner.
func NewChannelThrough[T any](sender ur.SenderRunner[T], receivers ...Receiver[T]) *ChannelThrough[T] {
	receivers = gcslc.SliceFilter(receivers, func(receiver Receiver[T]) bool {
		return receiver != nil
	})

	return &ChannelThrough[T]{
		receivers: receivers,
		sender:    sender,
	}
}

// Run is a method that runs the handler.
func (ct *ChannelThrough[T]) Run() {
	if ct.IsRunning() || len(ct.receivers) == 0 {
		return
	}

	if ct.sender == nil {
		go func() {
			defer ct.isClosed.Set(true)

			var wg sync.WaitGroup

			wg.Add(len(ct.receivers))

			for _, receiver := range ct.receivers {
				go func(receiver Receiver[T]) {
					defer wg.Done()

					DiscardAnyMessage(receiver)
				}(receiver)
			}

			wg.Wait()
		}()

		return
	}

	ct.buffer = NewBuffer[T]()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		ct.sender.Start()
	}()

	go func() {
		defer wg.Done()

		ct.buffer.Start()
	}()

	wg.Wait()

	go func() {
		defer ct.isClosed.Set(true)

		abruptClose := false

		for !abruptClose {
			msg, ok := ct.buffer.Receive()
			if !ok {
				break
			}

			ok = ct.sender.Send(msg)
			if !ok {
				abruptClose = true
			}
		}

		if abruptClose {
			DiscardAnyMessage(ct.buffer)
		} else {
			ct.sender.Close()
		}
	}()

	go func() {
		defer ct.buffer.Close()

		var wg sync.WaitGroup

		wg.Add(len(ct.receivers))

		for _, receiver := range ct.receivers {
			go func(receiver Receiver[T]) {
				defer wg.Done()

				for {
					msg, ok := receiver.Receive()
					if !ok {
						break
					}

					ok = ct.buffer.Send(msg)
					if !ok {
						panic("unable to send message to buffer")
					}
				}
			}(receiver)
		}

		wg.Wait()
	}()
}

// IsRunning is a method that returns true if the handler is running.
//
// Returns:
//   - bool: True if the handler is running, false otherwise.
func (ct *ChannelThrough[T]) IsRunning() bool {
	return ct.isClosed != nil && !ct.isClosed.Get()
}
