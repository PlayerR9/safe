package runner

import "fmt"

// ErrPanic represents an error when a panic occurs.
type ErrPanic struct {
	// Value is the value that caused the panic.
	Value any
}

// Error implements the error interface.
//
// Message: "panic: {value}"
func (e ErrPanic) Error() string {
	return fmt.Sprintf("panic: %v", e.Value)
}

// NewErrPanic creates a new ErrPanic error.
//
// Parameters:
//   - value: The value that caused the panic.
//
// Returns:
//   - *ErrPanic: A pointer to the newly created ErrPanic. Never returns nil.
func NewErrPanic(value any) *ErrPanic {
	return &ErrPanic{
		Value: value,
	}
}
