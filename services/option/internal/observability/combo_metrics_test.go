package observability

import (
	"reflect"
	"testing"

	clientprometheus "github.com/prometheus/client_golang/prometheus"
	coreprometheus "github.com/zeromicro/go-zero/core/prometheus"
)

type metricCall struct {
	value  float64
	labels []string
}

type fakeCounterVec struct {
	calls []metricCall
}

func (f *fakeCounterVec) Inc(labels ...string) {
	f.calls = append(f.calls, metricCall{value: 1, labels: append([]string(nil), labels...)})
}

type fakeGaugeVec struct {
	calls []metricCall
}

func (f *fakeGaugeVec) Set(value float64, labels ...string) {
	f.calls = append(f.calls, metricCall{value: value, labels: append([]string(nil), labels...)})
}

func TestComboMetricsUseBoundedOperationalLabels(t *testing.T) {
	isolation := &fakeCounterVec{}
	barrier := &fakeCounterVec{}
	stale := &fakeGaugeVec{}
	queryFailure := &fakeCounterVec{}
	originalIsolation := comboIsolationViolation
	originalBarrier := comboDebitBarrierViolation
	originalStale := comboDebitBarrierStaleEvents
	originalQueryFailure := comboObservabilityQueryFailure
	comboIsolationViolation = isolation
	comboDebitBarrierViolation = barrier
	comboDebitBarrierStaleEvents = stale
	comboObservabilityQueryFailure = queryFailure
	t.Cleanup(func() {
		comboIsolationViolation = originalIsolation
		comboDebitBarrierViolation = originalBarrier
		comboDebitBarrierStaleEvents = originalStale
		comboObservabilityQueryFailure = originalQueryFailure
	})

	RecordComboIsolationViolation(9, "public_book")
	RecordComboDebitBarrierViolation(9)
	SetComboDebitBarrierStaleEvents(0, 3)
	RecordComboObservabilityQueryFailure(0, "stale_debit_barrier")

	if !reflect.DeepEqual(isolation.calls, []metricCall{{
		value: 1, labels: []string{"9", "public_book"},
	}}) {
		t.Fatalf("isolation calls=%+v", isolation.calls)
	}
	if !reflect.DeepEqual(barrier.calls, []metricCall{{
		value: 1, labels: []string{"9"},
	}}) {
		t.Fatalf("barrier calls=%+v", barrier.calls)
	}
	if !reflect.DeepEqual(stale.calls, []metricCall{{
		value: 3, labels: []string{"all"},
	}}) {
		t.Fatalf("stale calls=%+v", stale.calls)
	}
	if !reflect.DeepEqual(queryFailure.calls, []metricCall{{
		value: 1, labels: []string{"all", "stale_debit_barrier"},
	}}) {
		t.Fatalf("query failure calls=%+v", queryFailure.calls)
	}
}

func TestComboMetricsAreRegisteredWithDocumentedNames(t *testing.T) {
	coreprometheus.Enable()
	RecordComboIsolationViolation(987654321, "registration_test")
	RecordComboDebitBarrierViolation(987654321)
	SetComboDebitBarrierStaleEvents(987654321, 7)
	RecordComboObservabilityQueryFailure(987654321, "registration_test")
	optionOperationsCount.Set(1, "987654321", "registration_test")
	optionOperationsOldestTimestamp.Set(1, "987654321", "registration_test")
	optionOperationsAmount.Set(1, "987654321", "registration_test", "USDT")
	optionOperationsSampleSuccess.Set(1)
	optionOperationsLastSuccessTimestamp.Set(1)
	optionOperationsSampleFailure.Inc("registration_test")
	optionRiskScanGroups.Set(1, "987654321")
	optionRiskScanFailedGroups.Set(1, "987654321")
	optionRiskScanFailureRatio.Set(1, "987654321")
	optionRiskScanLastCompletedTimestamp.Set(1, "987654321")
	optionRiskScanExecutionFailure.Inc("987654321", "registration_test")
	RecordAdminRejectedMutation(987654321, "contract", "listed_economics")

	families, err := clientprometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather prometheus metrics: %v", err)
	}
	want := map[string]bool{
		"wklive_option_combo_isolation_violation_total":            false,
		"wklive_option_combo_debit_barrier_violation_total":        false,
		"wklive_option_combo_debit_barrier_stale_events":           false,
		"wklive_option_combo_observability_query_failure_total":    false,
		"wklive_option_operations_count":                           false,
		"wklive_option_operations_oldest_timestamp_seconds":        false,
		"wklive_option_operations_amount":                          false,
		"wklive_option_operations_sample_success":                  false,
		"wklive_option_operations_last_success_timestamp_seconds":  false,
		"wklive_option_operations_sample_failure_total":            false,
		"wklive_option_risk_scan_groups":                           false,
		"wklive_option_risk_scan_failed_groups":                    false,
		"wklive_option_risk_scan_failure_ratio":                    false,
		"wklive_option_risk_scan_last_completed_timestamp_seconds": false,
		"wklive_option_risk_scan_execution_failure_total":          false,
		"wklive_option_admin_rejected_mutation_total":              false,
	}
	for _, family := range families {
		if _, ok := want[family.GetName()]; ok {
			want[family.GetName()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("prometheus metric %s is not registered", name)
		}
	}
}
