package tasks

import (
	"testing"

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
