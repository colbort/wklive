package tasklogic

import (
	"context"
	"errors"
	"strings"
	"testing"

	"wklive/common/alert"
	"wklive/common/notify"
	"wklive/services/trade/models"
)

type reconciliationAlertNotifierFake struct {
	value alert.Alert
	err   error
}

func (f *reconciliationAlertNotifierFake) Notify(
	_ context.Context,
	value alert.Alert,
) error {
	f.value = value
	return f.err
}

type reconciliationAlertReleaserFake struct {
	calls      int
	tenantID   int64
	issueKey   string
	reservedAt int64
	err        error
}

func (f *reconciliationAlertReleaserFake) ReleaseAlertReservation(
	_ context.Context,
	tenantID int64,
	issueKey string,
	reservedAt int64,
) error {
	f.calls++
	f.tenantID = tenantID
	f.issueKey = issueKey
	f.reservedAt = reservedAt
	return f.err
}

func TestDeliverReservedReconciliationAlertPublishesEvent(t *testing.T) {
	notifier := &reconciliationAlertNotifierFake{}
	releaser := &reconciliationAlertReleaserFake{}
	issue := reconciliationAlertTestIssue()

	if err := deliverReservedReconciliationAlert(
		context.Background(),
		notifier,
		releaser,
		issue,
	); err != nil {
		t.Fatal(err)
	}
	if notifier.value.Type != alert.TypeContractReconciliation ||
		notifier.value.Severity != notify.EventLevelError ||
		notifier.value.TenantID != issue.TenantId ||
		notifier.value.Key != issue.IssueKey {
		t.Fatalf("alert=%+v", notifier.value)
	}
	if releaser.calls != 0 {
		t.Fatalf("unexpected reservation release calls=%d", releaser.calls)
	}
}

func TestDeliverReservedReconciliationAlertReleasesFailedPublish(t *testing.T) {
	notifier := &reconciliationAlertNotifierFake{err: errors.New("kafka unavailable")}
	releaser := &reconciliationAlertReleaserFake{}
	issue := reconciliationAlertTestIssue()

	err := deliverReservedReconciliationAlert(
		context.Background(),
		notifier,
		releaser,
		issue,
	)
	if err == nil || !strings.Contains(err.Error(), "kafka unavailable") {
		t.Fatalf("err=%v", err)
	}
	if releaser.calls != 1 ||
		releaser.tenantID != issue.TenantId ||
		releaser.issueKey != issue.IssueKey ||
		releaser.reservedAt != issue.LastSeenAt {
		t.Fatalf("releaser=%+v", releaser)
	}
}

func TestDeliverReservedReconciliationAlertReportsReleaseFailure(t *testing.T) {
	notifier := &reconciliationAlertNotifierFake{err: errors.New("kafka unavailable")}
	releaser := &reconciliationAlertReleaserFake{err: errors.New("database unavailable")}

	err := deliverReservedReconciliationAlert(
		context.Background(),
		notifier,
		releaser,
		reconciliationAlertTestIssue(),
	)
	if err == nil ||
		!strings.Contains(err.Error(), "kafka unavailable") ||
		!strings.Contains(err.Error(), "database unavailable") {
		t.Fatalf("err=%v", err)
	}
}

func TestDeliverReservedReconciliationAlertRejectsNilIssue(t *testing.T) {
	err := deliverReservedReconciliationAlert(
		context.Background(),
		&reconciliationAlertNotifierFake{},
		&reconciliationAlertReleaserFake{},
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "issue is nil") {
		t.Fatalf("err=%v", err)
	}
}

func reconciliationAlertTestIssue() *models.TContractReconciliationIssue {
	return &models.TContractReconciliationIssue{
		TenantId:      900101,
		IssueKey:      "reservation:ACCEPT-ORDER-PERP-20260729-1",
		CheckType:     "reservation_balance",
		BizType:       "contract_order",
		BizNo:         "ACCEPT-ORDER-PERP-20260729-1",
		InstructionId: 990101,
		ExpectedValue: "100",
		ActualValue:   "99",
		Detail:        "acceptance fixture",
		LastSeenAt:    1785330000123,
		UpdateTimes:   1785330000123,
	}
}
