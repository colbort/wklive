package models

import "testing"

func TestReconciliationFindingAlertDue(t *testing.T) {
	finding := &TContractReconciliationIssue{
		CheckType: "ORDER_FILL", BizType: "order", BizNo: "ORDER-1",
		InstructionId: 7, ExpectedValue: "expected", ActualValue: "actual", Detail: "detail",
	}
	const (
		now      = int64(2_000_000)
		interval = int64(30 * 60 * 1000)
	)
	same := func() *TContractReconciliationIssue {
		return &TContractReconciliationIssue{
			Status: 1, LastAlertAt: now - interval + 1,
			CheckType: finding.CheckType, BizType: finding.BizType, BizNo: finding.BizNo,
			InstructionId: finding.InstructionId, ExpectedValue: finding.ExpectedValue,
			ActualValue: finding.ActualValue, Detail: finding.Detail,
		}
	}

	if !reconciliationFindingAlertDue(nil, finding, now, interval) {
		t.Fatal("new issue must alert")
	}
	if reconciliationFindingAlertDue(same(), finding, now, interval) {
		t.Fatal("unchanged issue inside reminder interval must be silent")
	}
	changed := same()
	changed.ActualValue = "old actual"
	if !reconciliationFindingAlertDue(changed, finding, now, interval) {
		t.Fatal("changed issue must alert immediately")
	}
	reminder := same()
	reminder.LastAlertAt = now - interval
	if !reconciliationFindingAlertDue(reminder, finding, now, interval) {
		t.Fatal("unchanged issue at reminder boundary must alert")
	}
	reopened := same()
	reopened.Status = 2
	if !reconciliationFindingAlertDue(reopened, finding, now, interval) {
		t.Fatal("reopened resolved issue must alert")
	}
	ignored := same()
	ignored.Status = 3
	ignored.ActualValue = "old actual"
	if reconciliationFindingAlertDue(ignored, finding, now, interval) {
		t.Fatal("manually ignored issue must stay silent")
	}
}
