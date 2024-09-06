package runner

// ErrAlreadyRunning is an error type that represents an error where
// a process is already running.
type ErrAlreadyRunning struct{}

// Error implements the error interface.
//
// Message: "the process is already running"
func (e *ErrAlreadyRunning) Error() string {
	return "the process is already running"
}

// NewErrAlreadyRunning creates a new ErrAlreadyRunning error.
//
// Returns:
//   - *ErrAlreadyRunning: The new error.
func NewErrAlreadyRunning() *ErrAlreadyRunning {
	e := &ErrAlreadyRunning{}
	return e
}
