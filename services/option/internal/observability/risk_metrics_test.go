package observability

import "testing"

func TestPublishRiskScanResultsPublishesExplicitTenantResultsWithoutClearingOthers(t *testing.T) {
	groups := &fakeGaugeVec{}
	failed := &fakeGaugeVec{}
	ratio := &fakeGaugeVec{}
	completed := &fakeGaugeVec{}
	originalGroups := optionRiskScanGroups
	originalFailed := optionRiskScanFailedGroups
	originalRatio := optionRiskScanFailureRatio
	originalCompleted := optionRiskScanLastCompletedTimestamp
	optionRiskScanGroups = groups
	optionRiskScanFailedGroups = failed
	optionRiskScanFailureRatio = ratio
	optionRiskScanLastCompletedTimestamp = completed
	riskScanMetricState.tenants = make(map[string]struct{})
	t.Cleanup(func() {
		optionRiskScanGroups = originalGroups
		optionRiskScanFailedGroups = originalFailed
		optionRiskScanFailureRatio = originalRatio
		optionRiskScanLastCompletedTimestamp = originalCompleted
		riskScanMetricState.tenants = make(map[string]struct{})
	})

	PublishRiskScanResults([]RiskScanTenantResult{{
		TenantID: 9, TotalGroups: 20, FailedGroups: 2,
	}}, 1000, 0)
	PublishRiskScanResults([]RiskScanTenantResult{{
		TenantID: 10, TotalGroups: 0, FailedGroups: 0,
	}}, 1010, 10)

	assertMetricCall(t, groups.calls, 0, 20, "9")
	assertMetricCall(t, failed.calls, 0, 2, "9")
	assertMetricCall(t, ratio.calls, 0, 0.1, "9")
	assertMetricCall(t, completed.calls, 0, 1000, "9")
	assertMetricCall(t, groups.calls, 1, 0, "10")
	assertMetricCall(t, failed.calls, 1, 0, "10")
	assertMetricCall(t, ratio.calls, 1, 0, "10")
	assertMetricCall(t, completed.calls, 1, 1010, "10")
	if len(groups.calls) != 2 || len(failed.calls) != 2 || len(ratio.calls) != 2 || len(completed.calls) != 2 {
		t.Fatalf("publishing tenant 10 unexpectedly rewrote another tenant: groups=%+v failed=%+v ratio=%+v completed=%+v",
			groups.calls, failed.calls, ratio.calls, completed.calls)
	}
}

func TestPublishRiskScanResultsGlobalScanClearsMissingTenant(t *testing.T) {
	groups := &fakeGaugeVec{}
	failed := &fakeGaugeVec{}
	ratio := &fakeGaugeVec{}
	completed := &fakeGaugeVec{}
	originalGroups := optionRiskScanGroups
	originalFailed := optionRiskScanFailedGroups
	originalRatio := optionRiskScanFailureRatio
	originalCompleted := optionRiskScanLastCompletedTimestamp
	optionRiskScanGroups = groups
	optionRiskScanFailedGroups = failed
	optionRiskScanFailureRatio = ratio
	optionRiskScanLastCompletedTimestamp = completed
	riskScanMetricState.tenants = make(map[string]struct{})
	t.Cleanup(func() {
		optionRiskScanGroups = originalGroups
		optionRiskScanFailedGroups = originalFailed
		optionRiskScanFailureRatio = originalRatio
		optionRiskScanLastCompletedTimestamp = originalCompleted
		riskScanMetricState.tenants = make(map[string]struct{})
	})

	PublishRiskScanResults([]RiskScanTenantResult{{
		TenantID: 9, TotalGroups: 20, FailedGroups: 2,
	}}, 1000, 0)
	PublishRiskScanResults(nil, 1010, 0)

	assertMetricCall(t, groups.calls, 1, 0, "9")
	assertMetricCall(t, failed.calls, 1, 0, "9")
	assertMetricCall(t, ratio.calls, 1, 0, "9")
	assertMetricCall(t, completed.calls, 1, 1010, "9")
}

func TestRecordRiskScanExecutionFailureUsesBoundedLabels(t *testing.T) {
	counter := &fakeCounterVec{}
	original := optionRiskScanExecutionFailure
	optionRiskScanExecutionFailure = counter
	t.Cleanup(func() {
		optionRiskScanExecutionFailure = original
	})

	RecordRiskScanExecutionFailure(9, "collect")
	if len(counter.calls) != 1 ||
		len(counter.calls[0].labels) != 2 ||
		counter.calls[0].labels[0] != "9" ||
		counter.calls[0].labels[1] != "collect" {
		t.Fatalf("risk scan execution failure calls=%+v", counter.calls)
	}
}
