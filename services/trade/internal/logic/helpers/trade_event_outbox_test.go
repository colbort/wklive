package helpers

import (
	"testing"
	"time"
)

func TestTradeEventRetryDelay(t *testing.T) {
	tests := []struct {
		retry int64
		want  time.Duration
	}{
		{retry: 0, want: time.Second},
		{retry: 1, want: time.Second},
		{retry: 2, want: 2 * time.Second},
		{retry: 10, want: 512 * time.Second},
		{retry: 20, want: 512 * time.Second},
	}
	for _, tt := range tests {
		if got := TradeEventRetryDelay(tt.retry); got != tt.want {
			t.Fatalf("retry %d: got %s, want %s", tt.retry, got, tt.want)
		}
	}
}
