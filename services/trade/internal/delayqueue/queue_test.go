package delayqueue

import (
	"testing"
	"time"
)

func TestRoundUpBeanstalkDelay(t *testing.T) {
	tests := []struct {
		name  string
		input time.Duration
		want  time.Duration
	}{
		{name: "past", input: -time.Millisecond, want: 0},
		{name: "now", input: 0, want: 0},
		{name: "subsecond", input: time.Millisecond, want: time.Second},
		{name: "exact second", input: 30 * time.Second, want: 30 * time.Second},
		{name: "fractional second", input: 29*time.Second + time.Millisecond, want: 30 * time.Second},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := roundUpBeanstalkDelay(test.input); got != test.want {
				t.Fatalf("roundUpBeanstalkDelay(%s)=%s, want %s", test.input, got, test.want)
			}
		})
	}
}
