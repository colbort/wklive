package tasks

import (
	"testing"
	"time"

	"wklive/services/option/internal/config"
)

func TestResolveMarketSnapshotInboxCleanupSettings(t *testing.T) {
	var c config.Config
	settings, adjusted := resolveMarketSnapshotInboxCleanupSettings(c)
	if adjusted {
		t.Fatal("default retention should not be reported as adjusted")
	}
	if settings.retention != 30*24*time.Hour || settings.interval != time.Hour ||
		settings.batchSize != 5000 || settings.maxBatches != 10 {
		t.Fatalf("unexpected defaults: %+v", settings)
	}

	c.MarketSnapshotInboxCleanup.RetentionHours = 24
	settings, adjusted = resolveMarketSnapshotInboxCleanupSettings(c)
	if !adjusted {
		t.Fatal("unsafe retention should be adjusted")
	}
	if settings.retention != minimumMarketSnapshotInboxRetention {
		t.Fatalf("retention = %s", settings.retention)
	}
}
