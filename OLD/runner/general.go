package runner

import (
	"context"
	"iter"
	"sync"
)

// Handler is a struct that represents a Go routine handler.
type Handler[O any] struct {
	// Data is the result of the Go routine.
	Data O

	// Err is the error status of the Go routine.
	Err error
}

// DoFunc is a function that defines the behavior of a Go routine.
//
// Parameters:
//   - ctx: The context of the Go routine.
//   - elem: The element to process.
//
// Returns:
//   - O: The result of the Go routine.
//   - error: The error status of the Go routine.
type DoFunc[I, O any] func(ctx context.Context, elem I) (O, error)

// ExecuteBatch executes a batch of Go routines.
//
// Parameters:
//   - ctx: The context of the batch.
//   - elems: The elements to process.
//   - do_fn: The function that defines the behavior of the Go routines.
//
// Returns:
//   - []Handler[O]: The results of the Go routines.
func ExecuteBatch[I, O any](ctx context.Context, elems iter.Seq[I], do_fn DoFunc[I, O]) []Handler[O] {
	if elems == nil || do_fn == nil {
		return nil
	}

	var handlers []Handler[O]
	var mu sync.Mutex

	var wg sync.WaitGroup

	for elem := range elems {
		wg.Add(1)

		go func(elem I) {
			defer wg.Done()

			output, err := do_fn(ctx, elem)

			h := Handler[O]{
				Data: output,
				Err:  err,
			}

			mu.Lock()
			defer mu.Unlock()

			handlers = append(handlers, h)
		}(elem)
	}

	wg.Wait()

	return handlers
}
