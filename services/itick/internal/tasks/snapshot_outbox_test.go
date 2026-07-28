package tasks

import (
	"testing"
	"time"

	"wklive/services/itick/models"
)

func TestSnapshotOutboxUnhealthy(t *testing.T) {
	for _, tc := range []struct {
		name   string
		health *models.SnapshotOutboxHealth
		age    int64
		want   bool
	}{
		{name: "healthy empty", health: &models.SnapshotOutboxHealth{}},
		{name: "fresh pending", health: &models.SnapshotOutboxHealth{PendingCount: 1}, age: 30_000},
		{name: "stale pending", health: &models.SnapshotOutboxHealth{PendingCount: 1}, age: 60_001, want: true},
		{name: "failed", health: &models.SnapshotOutboxHealth{FailedCount: 1}, want: true},
		{name: "manual", health: &models.SnapshotOutboxHealth{ManualCount: 1}, want: true},
		{name: "missing health", health: nil, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := snapshotOutboxUnhealthy(tc.health, tc.age); got != tc.want {
				t.Fatalf("unhealthy=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestSnapshotOutboxDrainMetrics(t *testing.T) {
	now := time.Unix(1000, 0)
	previous := &snapshotOutboxHealthSample{at: now, openRows: 1000}
	current := &snapshotOutboxHealthSample{at: now.Add(10 * time.Second), openRows: 800}
	rate, eta := snapshotOutboxDrainMetrics(previous, current)
	if rate != 20 || eta != 40 {
		t.Fatalf("rate=%v eta=%d want=20/40", rate, eta)
	}
	current.openRows = 1100
	rate, eta = snapshotOutboxDrainMetrics(previous, current)
	if rate != 0 || eta != -1 {
		t.Fatalf("growing backlog rate=%v eta=%d want=0/-1", rate, eta)
	}
}
