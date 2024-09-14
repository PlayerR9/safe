package runner

import (
	"errors"
	"sync"

	rws "github.com/PlayerR9/safe/rw_safe"
)

var (
	// AlreadyRunning is the error that is returned when the process is already running.
	AlreadyRunning error
)

func init() {
	AlreadyRunning = errors.New("the process is already running")
}

// Batch is a struct that represents a batch of Go routines.
type Batch struct {
	// handlers is a map of the identifiers of the Go routines to the GRHandler
	// instances that handle them.
	handlers map[string]*HandlerSimple
}

// NewBatch creates a new batch of Go routines.
//
// Returns:
//   - *Batch: The new batch. Never returns nil.
func NewBatch() *Batch {
	return &Batch{
		handlers: make(map[string]*HandlerSimple),
	}
}

// Add is a method of Batch that adds a Go routine to the batch.
//
// Parameters:
//   - identifier: The identifier of the Go routine.
//   - routine: The Go routine to add to the batch.
//
// Behaviors:
//   - It ignores nil Go routines.
//   - It replaces the Go routine if the identifier already exists in the batch.
func (b *Batch) Add(identifier string, routine func() error) {
	if b == nil || routine == nil {
		return
	}

	h, _ := NewHandlerSimple(routine)

	b.handlers[identifier] = h
}

// Clear is a method of Batch that clears the batch.
func (b *Batch) Clear() {
	if b == nil {
		return
	}

	if len(b.handlers) > 0 {
		for k := range b.handlers {
			b.handlers[k] = nil

			delete(b.handlers, k)
		}
	}

	b.handlers = make(map[string]*HandlerSimple)
}

// StartAll is a function that starts all Go routines in the batch.
//
// Parameters:
//   - batch: A slice of pointers to the GRHandler instances that handle the Go routines.
func (b *Batch) StartAll() {
	if b == nil || len(b.handlers) == 0 {
		return
	}

	for _, h := range b.handlers {
		h.Start()
	}
}

// WaitAll is a function that waits for all Go routines in the batch to finish
// and returns a slice of errors that represent the error statuses of the Go routines.
//
// Parameters:
//   - batch: A slice of pointers to the GRHandler instances that handle the Go routines.
//
// Returns:
//   - map[string]error: A map of the error statuses of the Go routines.
func (b *Batch) WaitAll() map[string]error {
	if b == nil || len(b.handlers) == 0 {
		return nil
	}

	errMap := rws.NewSafeMap[string, error]()

	for k := range b.handlers {
		errMap.Set(k, nil)
	}

	var wg sync.WaitGroup

	wg.Add(len(b.handlers))

	for k, h := range b.handlers {
		go func(k string, h *HandlerSimple) {
			defer wg.Done()

			for {
				err, ok := h.ReceiveErr()
				if !ok {
					return
				}

				errMap.Set(k, err)

				if err != nil {
					return
				}
			}
		}(k, h)
	}

	wg.Wait()

	m := errMap.GetMap()
	return m
}
