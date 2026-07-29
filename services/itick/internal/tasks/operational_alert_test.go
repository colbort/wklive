package tasks

import (
	"context"
	"errors"
	"testing"

	"wklive/common/alert"
	"wklive/common/notify"
	"wklive/services/itick/internal/priceengine"
	"wklive/services/itick/models"
)

type capturedAlertNotifier struct {
	value alert.Alert
	err   error
}

func (n *capturedAlertNotifier) Notify(_ context.Context, value alert.Alert) error {
	n.value = value
	return n.err
}

func TestPublishPriceEngineInputAlert(t *testing.T) {
	notifier := &capturedAlertNotifier{}
	if err := publishPriceEngineInputAlert(
		context.Background(),
		notifier,
		"firing",
		"missing Binance input",
		1234,
	); err != nil {
		t.Fatal(err)
	}
	if notifier.value.Type != alert.TypePriceEngineInput ||
		notifier.value.Key != "price-engine-input" ||
		notifier.value.Severity != notify.EventLevelError ||
		notifier.value.State != alert.StateFiring {
		t.Fatalf("alert=%+v", notifier.value)
	}
}

func TestPublishSnapshotOutboxHealthAlert(t *testing.T) {
	notifier := &capturedAlertNotifier{}
	health := &models.SnapshotOutboxHealth{PendingCount: 12, FailedCount: 1}
	if err := publishSnapshotOutboxHealthAlert(
		context.Background(),
		notifier,
		"resolved",
		"snapshot-outbox",
		"healthy",
		health,
		0,
		-1,
		1234,
	); err != nil {
		t.Fatal(err)
	}
	if notifier.value.Type != alert.TypeSnapshotOutbox ||
		notifier.value.State != alert.StateResolved ||
		notifier.value.Key != "snapshot-outbox" {
		t.Fatalf("alert=%+v", notifier.value)
	}
}

func TestPriceEngineInputAlertFingerprintExcludesTarget(t *testing.T) {
	first := &priceengine.InputUnavailableError{
		FormulaNo:    "BTCUSDT-INDEX-v1",
		Authority:    "binance-public",
		Kind:         "FINAL_QUOTE",
		CategoryCode: "crypto",
		Market:       "BINANCE",
		Symbol:       "BTCUSDT",
		Target:       1000,
		Detail:       "sql: no rows in result set",
	}
	second := *first
	second.Target = 2000
	if first.Error() == second.Error() {
		t.Fatal("operator-facing errors must retain target time")
	}
	if priceEngineInputAlertFingerprint(first) != priceEngineInputAlertFingerprint(&second) {
		t.Fatal("alert fingerprint must ignore target time")
	}
	if !errors.Is(first, priceengine.ErrInputUnavailable) {
		t.Fatal("typed error must preserve ErrInputUnavailable classification")
	}
}
