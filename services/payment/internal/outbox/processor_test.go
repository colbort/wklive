package outbox

import (
	"strings"
	"testing"
	"time"
)

func TestRetryDelayIsBounded(t *testing.T) {
	tests := []struct {
		retry int64
		want  time.Duration
	}{
		{retry: 0, want: time.Second},
		{retry: 1, want: time.Second},
		{retry: 10, want: 10 * time.Second},
		{retry: 60, want: time.Minute},
		{retry: 600, want: time.Minute},
	}
	for _, test := range tests {
		if got := retryDelay(test.retry); got != test.want {
			t.Fatalf("retry=%d delay=%v want=%v", test.retry, got, test.want)
		}
	}
}

func TestTruncateErrorUsesRuneLimit(t *testing.T) {
	message := strings.Repeat("错", 1_001)
	got := truncateError(message, 1_000)
	if len([]rune(got)) != 1_000 {
		t.Fatalf("truncated runes=%d want=1000", len([]rune(got)))
	}
}
