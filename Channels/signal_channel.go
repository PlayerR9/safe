package Channels

import "sync"

// DoFunc is a function that is called when a signal is received.
//
// Parameters:
//   - code: The code of the signal.
type DoFunc func(code int)

// SignalChannel is a channel for signals.
type SignalChannel struct {
	// signalChan is the channel for signals.
	signalChan chan int

	// doFunc is the function to call when a signal is received.
	doFunc DoFunc

	// wg is the wait group for the SignalChannel.
	wg sync.WaitGroup
}

// Start starts the SignalChannel.
func (sc *SignalChannel) Start() {
	sc.wg.Add(1)

	go sc.signalListener()
}

// Close closes the SignalChannel.
func (sc *SignalChannel) Close() {
	close(sc.signalChan)

	sc.wg.Wait()

	sc.signalChan = nil
}

// Wait waits for the SignalChannel to finish.
func (sc *SignalChannel) Wait() {
	sc.wg.Wait()
}

// NewSignalChannel creates a new SignalChannel.
//
// Parameters:
//   - doFunc: The function to call when a signal is received.
//
// Returns:
//   - *SignalChannel: The new SignalChannel.
//
// Behaviors:
//   - If doFunc is nil, an empty function is used.
func NewSignalChannel(doFunc DoFunc) *SignalChannel {
	if doFunc == nil {
		doFunc = func(code int) {}
	}

	return &SignalChannel{
		signalChan: make(chan int),
		doFunc:     doFunc,
	}
}

// signalListener is a helper function that listens for signals.
func (sc *SignalChannel) signalListener() {
	defer sc.wg.Done()

	for val := range sc.signalChan {
		sc.doFunc(val)
	}
}

// Send sends a signal to the SignalChannel.
//
// Parameters:
//   - code: The code of the signal.
func (sc *SignalChannel) Send(code int) {
	sc.signalChan <- code
}
