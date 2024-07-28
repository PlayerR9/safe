package Buffer

import (
	"sync"
	"testing"
)

func TestInit(t *testing.T) {
	const (
		MaxCount int = 100
	)

	buffer := NewBuffer[int]()

	buffer.Start()

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		for i := 0; i < MaxCount; i++ {
			ok := buffer.Send(i)
			if !ok {
				t.Errorf("could not send %d", i)
				return
			}
		}

		buffer.Close()
	}()

	go func() {
		defer wg.Done()

		i := 0

		for i < MaxCount {
			x, ok := buffer.Receive()
			if !ok {
				t.Errorf("could not receive %d", i)
				return
			}

			if x != i {
				t.Errorf("expected %d, got %d", i, x)
				return
			}

			i++
		}
	}()

	wg.Wait()

	t.Fatalf("Test completed")
}

func TestTrimFrom(t *testing.T) {
	const (
		MaxCount int = 100
	)

	buffer := NewBuffer[int]()

	var wg sync.WaitGroup

	wg.Add(1)

	buffer.Start()

	go func(max int) {
		defer wg.Done()

		for {
			x, ok := buffer.Receive()
			if !ok {
				break
			}

			t.Logf("Received %d", x)
		}
	}(MaxCount)

	for i := 0; i < MaxCount; i++ {
		ok := buffer.Send(i)
		if !ok {
			t.Errorf("Could not send %d", i)
			return
		}
	}

	buffer.CleanBuffer()

	buffer.Close()

	wg.Wait()

	_, ok := buffer.Receive()
	if ok {
		t.Errorf("Expected false, got %t", ok)
	}
}
