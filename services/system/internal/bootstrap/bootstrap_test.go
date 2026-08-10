package bootstrap

import (
	"errors"
	"testing"
)

func TestRetryEventuallySucceeds(t *testing.T) {
	attempts := 0
	err := retry(3, 0, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("database unavailable")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("attempts=%d want=3", attempts)
	}
}

func TestRetryReturnsLastError(t *testing.T) {
	want := errors.New("database unavailable")
	attempts := 0
	err := retry(2, 0, func() error {
		attempts++
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("error=%v want wrapped %v", err, want)
	}
	if attempts != 2 {
		t.Fatalf("attempts=%d want=2", attempts)
	}
}

func TestRetryRejectsInvalidInput(t *testing.T) {
	if err := retry(0, 0, func() error { return nil }); err == nil {
		t.Fatal("expected invalid attempts error")
	}
	if err := retry(1, 0, nil); err == nil {
		t.Fatal("expected nil function error")
	}
}
