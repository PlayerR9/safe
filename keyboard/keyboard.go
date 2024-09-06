package keyboard

import (
	"context"
	_ "image/png"
	"sync"

	sfb "github.com/PlayerR9/safe/buffer"
	"github.com/eiannone/keyboard"
)

// Keyboard handles keyboard input using the eiannone/keyboard package.
type Keyboard struct {
	// buffer is a safe buffer for keyboard.Key values.
	buffer *sfb.Buffer[keyboard.Key]

	// errChan is the error channel for the Keyboard.
	errChan chan error

	// ctx is the context for the Keyboard.
	ctx context.Context

	// cancel is the cancel function for the Keyboard.
	cancel context.CancelFunc

	// wg is the wait group for the Keyboard.
	wg sync.WaitGroup
}

// NewKeyboard creates a new Keyboard.
//
// Returns:
//   - *Keyboard: The new Keyboard.
func NewKeyboard() *Keyboard {
	k := &Keyboard{}

	k.buffer = sfb.NewBuffer[keyboard.Key]()

	return k
}

// GetErrorChannel returns the error channel for the Keyboard.
//
// Returns:
//   - <-chan error: The error channel.
func (k *Keyboard) GetErrorChannel() <-chan error {
	return k.errChan
}

// GetKeyChannel returns the key channel for the Keyboard.
//
// Returns:
//   - <-chan keyboard.Key: The key channel.
func (k *Keyboard) GetKeyReceiver() sfb.Receiver[keyboard.Key] {
	return k.buffer
}

// Close closes the Keyboard.
//
// Returns:
//   - error: An error if the Keyboard could not be closed.
func (k *Keyboard) Close() {
	select {
	case <-k.ctx.Done():
		// Do nothing as the Keyboard is already closed.
	default:
		k.buffer.Close()

		k.cancel()

		k.wg.Wait()

		close(k.errChan)

		// Clean up

		k.buffer = nil

		err := keyboard.Close()
		if err != nil {
			panic(err)
		}
	}
}

// Start starts the Keyboard.
//
// Returns:
//   - error: An error if the Keyboard could not be started.
func (k *Keyboard) Start() error {
	select {
	case <-k.ctx.Done():
		k.ctx, k.cancel = context.WithCancel(context.Background())

		k.errChan = make(chan error)

		err := keyboard.Open()
		if err != nil {
			return err
		}

		k.wg.Add(1)

		go k.keyListener()
	default:
		// Do nothing as the Keyboard is already started.
	}

	return nil
}

// Wait waits for the Keyboard to finish.
// It may cause a deadlock if the Keyboard is not closed.
func (k *Keyboard) Wait() {
	k.wg.Wait()
}

// keyListener is an helper function that listens for keyboard input.
func (k *Keyboard) keyListener() {
	defer k.wg.Done()

	for {
		select {
		case <-k.ctx.Done():
			return
		default:
			_, key, err := keyboard.GetKey()
			if err != nil {
				k.errChan <- err
			} else {
				k.buffer.Send(key)
			}
		}
	}
}
