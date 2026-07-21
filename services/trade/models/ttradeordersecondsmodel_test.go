package models

import "testing"

func TestSecondsRetryState(t *testing.T) {
	const (
		status = int64(3)
		now    = int64(1_000_000)
	)

	tests := []struct {
		name       string
		retry      int64
		wantStatus int64
		wantAt     int64
		wantManual bool
	}{
		{name: "first retry", retry: 1, wantStatus: status, wantAt: now + 2_000},
		{name: "delay capped", retry: 19, wantStatus: status, wantAt: now + 1_024_000},
		{name: "manual threshold", retry: 20, wantStatus: 7, wantAt: 0, wantManual: true},
		{name: "past threshold", retry: 21, wantStatus: 7, wantAt: 0, wantManual: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotAt, gotManual := secondsRetryState(status, tt.retry, now)
			if gotStatus != tt.wantStatus || gotAt != tt.wantAt || gotManual != tt.wantManual {
				t.Fatalf("secondsRetryState() = (%d, %d, %t), want (%d, %d, %t)", gotStatus, gotAt, gotManual, tt.wantStatus, tt.wantAt, tt.wantManual)
			}
		})
	}
}
