package runner

import (
	"context"
	"sync"

	gcers "github.com/PlayerR9/go-commons/errors"
)

// HandlerSimple is a struct that represents a Go routine handler.
// It is used to handle the result of a Go routine.
type HandlerSimple struct {
	// wg is a WaitGroup that is used to wait for the Go routine to finish.
	wg sync.WaitGroup

	// errChan is the error status of the Go routine.
	errChan chan error

	// routine is the Go routine that is run by the handler.
	routine func() error

	// ctx is the context of the Go routine.
	ctx context.Context

	// cancel is the cancel function of the Go routine.
	cancel context.CancelFunc
}

// Start implements the Runner interface.
func (h *HandlerSimple) Start() {
	h.errChan = make(chan error)

	h.ctx, h.cancel = context.WithCancel(context.Background())

	h.wg.Add(1)

	go h.run()

}

// Close implements the Runner interface.
func (h *HandlerSimple) Close() {
	select {
	case <-h.ctx.Done():
		// Do nothing as the context is already done.
	default:
		h.cancel()

		h.wg.Wait()
	}
}

// IsClosed implements the Runner interface.
func (h *HandlerSimple) IsClosed() bool {
	return h.errChan == nil
}

// ReceiveErr implements the Runner interface.
func (h *HandlerSimple) ReceiveErr() (error, bool) {
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

// run is a private method of HandlerSimple that is runned by the Go routine.
//
// Behaviors:
//   - Use uc.ErrNoError to exit the Go routine as nil is used to signal
//     that the function has finished successfully but the Go routine is still running.
func (h *HandlerSimple) run() {
	defer h.wg.Done()

	defer func() {
		r := recover()

		if r != nil {
			h.errChan <- gcers.NewErrPanic(r)
		}

		h.clean()
	}()

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			err := h.routine()
			if err != nil {
				h.errChan <- err
				return
			}
		}
	}
}

// NewHandlerSimple creates a new HandlerSimple.
//
// Parameters:
//   - routine: The Go routine to run.
//
// Returns:
//   - *HandlerSimple: A pointer to the HandlerSimple that handles the result of the Go routine.
//
// Behaviors:
//   - If routine is nil, this function returns nil.
//   - The Go routine is not started automatically.
//   - In routine, use *uc.ErrNoError to exit the Go routine as nil is used to signal
//     that the function has finished successfully but the Go routine is still running.
func NewHandlerSimple(routine func() error) *HandlerSimple {
	if routine == nil {
		return nil
	}

	hs := &HandlerSimple{
		routine: routine,
	}

	return hs
}

// clean is a private method of HandlerSimple that cleans up the handler.
func (h *HandlerSimple) clean() {
	if h.errChan != nil {
		close(h.errChan)
		h.errChan = nil
	}
}
