package helpers

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestSubscriberRestartsUntilContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32
	done := make(chan struct{})
	go func() {
		defer close(done)
		runSubscriberWithRestart(ctx, "acceptance subscriber", time.Millisecond, 4*time.Millisecond, func() error {
			if calls.Add(1) == 3 {
				cancel()
			}
			return errors.New("broker unavailable")
		})
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("subscriber restart loop did not stop")
	}
	if calls.Load() != 3 {
		t.Fatalf("subscribe calls=%d want=3", calls.Load())
	}
}

func TestSubscriberRestartRejectsInvalidConfiguration(t *testing.T) {
	var calls atomic.Int32
	runSubscriberWithRestart(context.Background(), "invalid", 0, time.Second, func() error {
		calls.Add(1)
		return nil
	})
	if calls.Load() != 0 {
		t.Fatal("invalid restart configuration invoked subscriber")
	}
}
