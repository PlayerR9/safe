package buffer

import (
	"fmt"
	"log"
	"strings"
	"sync"
)

// Debugger is a struct that provides a way to print debug messages.
type Debugger struct {
	// logger is the logger to use.
	logger *log.Logger

	// msgBuffer is the buffer for messages.
	msgBuffer *Buffer[string]

	// debugMode is the flag that determines whether or not to print debug messages.
	debugMode bool

	// wg is the wait group for the goroutines.
	wg sync.WaitGroup
}

// Start implements the Runner interface.
func (d *Debugger) Start() {
	if d.msgBuffer != nil {
		return
	}

	d.msgBuffer = NewBuffer[string]()
	d.msgBuffer.Start()

	d.wg.Add(1)

	if d.logger == nil {
		go d.stdoutListener()
	} else {
		go d.loggerListener()
	}
}

// Close implements the Runner interface.
func (d *Debugger) Close() {
	if d.msgBuffer == nil {
		// Already closed
		return
	}

	d.msgBuffer.Close()

	d.wg.Wait()

	d.msgBuffer = nil
	d.logger = nil
}

// IsClosed returns true if the runner is closed, false otherwise.
func (d *Debugger) IsClosed() bool {
	return d.msgBuffer == nil
}

// NewDebugger is a function that creates a new debugger.
//
// Parameters:
//   - logger: The logger to use.
//
// Returns:
//   - *Debugger: The new debugger.
func NewDebugger(logger *log.Logger) *Debugger {
	return &Debugger{
		logger: logger,
	}
}

// ToggleDebugMode is a function that toggles the debug mode.
//
// Parameters:
//   - active: The flag to set the debug mode.
func (d *Debugger) ToggleDebugMode(active bool) {
	d.debugMode = active
}

// loggerListener is a function that listens for messages and logs them.
func (d *Debugger) loggerListener() {
	defer d.wg.Done()

	for {
		msg, ok := d.msgBuffer.Receive()
		if !ok {
			break
		}

		if strings.HasSuffix(msg, "\n") {
			d.logger.Print(msg)
		} else {
			d.logger.Println(msg)
		}
	}
}

// stdoutListener is a function that listens for messages and prints them to stdout.
func (d *Debugger) stdoutListener() {
	defer d.wg.Done()

	for {
		msg, ok := d.msgBuffer.Receive()
		if !ok {
			break
		}

		if strings.HasSuffix(msg, "\n") {
			fmt.Print(msg)
		} else {
			fmt.Println(msg)
		}
	}
}

// Println is a function that prints a line.
//
// Parameters:
//   - v: The values to print.
func (d *Debugger) Println(v ...interface{}) {
	if !d.debugMode || d.msgBuffer == nil {
		return
	}

	d.msgBuffer.Send(fmt.Sprintln(v...))
}

// Printf is a function that prints formatted text.
//
// '\n' is always appended to the end of the format string.
//
// Parameters:
//   - format: The format string.
//   - v: The values to print.
func (d *Debugger) Printf(format string, v ...interface{}) {
	if !d.debugMode || d.msgBuffer == nil {
		return
	}

	d.msgBuffer.Send(fmt.Sprintf(format, v...))
}

// Write is a function that writes to the debugger.
//
// '\n' is always appended to the end of the bytes.
//
// Parameters:
//   - p: The bytes to write.
//
// Returns:
//   - int: Always the length of the bytes.
//   - error: Always nil.
func (d *Debugger) Write(p []byte) (n int, err error) {
	if !d.debugMode || d.msgBuffer == nil {
		return 0, nil
	}

	d.msgBuffer.Send(string(p))

	return len(p), nil
}

// GetDebugMode is a function that returns the debug mode.
//
// Returns:
//   - bool: The debug mode.
func (d *Debugger) GetDebugMode() bool {
	return d.debugMode
}
