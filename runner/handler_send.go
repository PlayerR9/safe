package runner

import (
	"errors"
	"sync"

	uc "github.com/PlayerR9/go-commons/errors"
)

var (
	// NoError is the error that is returned when there is no error. Readers must return
	// this error as is an not wrap it as callers are expected to check for this error
	// with ==.
	NoError error
)

func init() {
	NoError = errors.New("no error")
}

// HandlerSend is a handler that, unlike HandlerSimple, whenever
// an error occurs, it is sent to the error channel instead of
// terminating the Go routine.
type HandlerSend[T any] struct {
	// wg is a WaitGroup that is used to wait for the Go routine to finish.
	wg sync.WaitGroup

	// errChan is the error status of the Go routine.
	errChan chan error

	// routine is the Go routine that is run by the handler.
	routine func(T) error

	// sendChan is the channel to send messages to the Go routine.
	sendChan chan T
}

// Start implements the Runner interface.
func (h *HandlerSend[T]) Start() {
	if h.sendChan != nil {
		return
	}

	h.errChan = make(chan error)
	h.sendChan = make(chan T)

	h.wg.Add(1)

	go h.run()
}

// Close implements the Runner interface.
func (h *HandlerSend[T]) Close() {
	if h.sendChan == nil {
		return
	}

	close(h.sendChan)
	h.sendChan = nil

	h.wg.Wait()

	h.clean()
}

// IsClosed implements the Runner interface.
func (h *HandlerSend[T]) IsClosed() bool {
	return h.errChan == nil
}

// ReceiveErr implements the Runner interface.
func (h *HandlerSend[T]) ReceiveErr() (error, bool) {
	if h.errChan == nil {
		return nil, false
	}

	err, ok := <-h.errChan
	if !ok {
		return nil, false
	} else {
		return err, true
	}
}

// run is a private method of HandlerSend that is runned by the Go routine.
//
// Behaviors:
//   - Use uc.ErrNoError to exit the Go routine as nil is used to signal
//     that the function has finished successfully but the Go routine is still running.
func (h *HandlerSend[T]) run() {
	defer h.wg.Done()

	defer func() {
		r := recover()

		if r != nil {
			h.errChan <- uc.NewErrPanic(r)
		}

		h.clean()
	}()

	for msg := range h.sendChan {
		err := h.routine(msg)
		if err == nil {
			continue
		} else if err == NoError {
			break
		}

		h.errChan <- err
	}
}

// NewHandlerSend creates a new HandlerSend.
//
// Parameters:
//   - routine: The Go routine to run.
//
// Returns:
//   - *HandlerSend: A pointer to the HandlerSend that handles the result of the Go routine.
//
// Behaviors:
//   - The Go routine is not started automatically.
//   - In routine, use *uc.ErrNoError to exit the Go routine as nil is used to signal
//     that the function has finished successfully but the Go routine is still running.
//   - If routine is nil, this function returns nil.
func NewHandlerSend[T any](routine func(T) error) *HandlerSend[T] {
	if routine == nil {
		return nil
	}

	hs := &HandlerSend[T]{
		routine: routine,
	}

	return hs
}

// Send is a method of HandlerSend that sends a message to the Go routine.
// If the Go routine is not running, false is returned.
//
// Parameters:
//   - msg: The message to send.
//
// Returns:
//   - bool: True if the message is sent, false otherwise.
func (h *HandlerSend[T]) Send(msg T) bool {
	if h.sendChan == nil {
		return false
	}

	h.sendChan <- msg

	return true
}

// clean is a private method of HandlerSend that cleans up the handler.
func (h *HandlerSend[T]) clean() {
	if h.errChan != nil {
		close(h.errChan)
		h.errChan = nil
	}

	if h.sendChan != nil {
		close(h.sendChan)
		h.sendChan = nil
	}
}
