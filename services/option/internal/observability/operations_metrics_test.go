package observability

import (
	"context"
	"errors"
	"testing"

	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestSampleOptionOperationsMetricsPublishesAndClearsTenantSeries(t *testing.T) {
	countGauge := &fakeGaugeVec{}
	oldestGauge := &fakeGaugeVec{}
	amountGauge := &fakeGaugeVec{}
	successGauge := &fakeGaugeVec{}
	lastSuccessGauge := &fakeGaugeVec{}
	failureCounter := &fakeCounterVec{}
	originalCount := optionOperationsCount
	originalOldest := optionOperationsOldestTimestamp
	originalAmount := optionOperationsAmount
	originalSuccess := optionOperationsSampleSuccess
	originalLastSuccess := optionOperationsLastSuccessTimestamp
	originalFailure := optionOperationsSampleFailure
	originalQuery := queryOptionOperationsMetrics
	optionOperationsCount = countGauge
	optionOperationsOldestTimestamp = oldestGauge
	optionOperationsAmount = amountGauge
	optionOperationsSampleSuccess = successGauge
	optionOperationsLastSuccessTimestamp = lastSuccessGauge
	optionOperationsSampleFailure = failureCounter
	operationsMetricsState.lastAttempt = 0
	operationsMetricsState.counts = make(map[operationsCountKey]struct{})
	operationsMetricsState.amounts = make(map[operationsAmountKey]struct{})
	queryCalls := 0
	queryOptionOperationsMetrics = func(
		context.Context, sqlx.SqlConn, int64, int64, int64,
	) ([]*models.OptionOperationsMetric, []*models.OptionOperationsAmountMetric, error) {
		queryCalls++
		if queryCalls == 1 {
			return []*models.OptionOperationsMetric{{
					TenantID: 9, Category: "asset_failed", Count: 2, Oldest: 800,
				}}, []*models.OptionOperationsAmountMetric{{
					TenantID: 9, Category: "unresolved_deficit", Coin: "USDT",
					Amount: decimal.RequireFromString("3.5"),
				}}, nil
		}
		return nil, nil, nil
	}
	t.Cleanup(func() {
		optionOperationsCount = originalCount
		optionOperationsOldestTimestamp = originalOldest
		optionOperationsAmount = originalAmount
		optionOperationsSampleSuccess = originalSuccess
		optionOperationsLastSuccessTimestamp = originalLastSuccess
		optionOperationsSampleFailure = originalFailure
		queryOptionOperationsMetrics = originalQuery
		operationsMetricsState.lastAttempt = 0
		operationsMetricsState.counts = make(map[operationsCountKey]struct{})
		operationsMetricsState.amounts = make(map[operationsAmountKey]struct{})
	})

	if err := SampleOptionOperationsMetrics(context.Background(), nil, 1000); err != nil {
		t.Fatalf("first sample: %v", err)
	}
	if queryCalls != 1 {
		t.Fatalf("query calls=%d want=1", queryCalls)
	}
	if err := SampleOptionOperationsMetrics(context.Background(), nil, 1001); err != nil {
		t.Fatalf("throttled sample: %v", err)
	}
	if queryCalls != 1 {
		t.Fatalf("throttled query calls=%d want=1", queryCalls)
	}
	if err := SampleOptionOperationsMetrics(context.Background(), nil, 1015); err != nil {
		t.Fatalf("clearing sample: %v", err)
	}
	assertMetricCall(t, countGauge.calls, 0, 2, "9", "asset_failed")
	assertMetricCall(t, oldestGauge.calls, 0, 800, "9", "asset_failed")
	assertMetricCall(t, amountGauge.calls, 0, 3.5, "9", "unresolved_deficit", "USDT")
	assertMetricCall(t, countGauge.calls, 1, 0, "9", "asset_failed")
	assertMetricCall(t, oldestGauge.calls, 1, 0, "9", "asset_failed")
	assertMetricCall(t, amountGauge.calls, 1, 0, "9", "unresolved_deficit", "USDT")
	assertMetricCall(t, successGauge.calls, 0, 1)
	assertMetricCall(t, successGauge.calls, 1, 1)
	assertMetricCall(t, lastSuccessGauge.calls, 0, 1000)
	assertMetricCall(t, lastSuccessGauge.calls, 1, 1015)
	if len(failureCounter.calls) != 0 {
		t.Fatalf("unexpected sample failures: %+v", failureCounter.calls)
	}
}

func TestSampleOptionOperationsMetricsReportsFailureAndRetainsPreviousSeries(t *testing.T) {
	successGauge := &fakeGaugeVec{}
	failureCounter := &fakeCounterVec{}
	originalSuccess := optionOperationsSampleSuccess
	originalFailure := optionOperationsSampleFailure
	originalQuery := queryOptionOperationsMetrics
	optionOperationsSampleSuccess = successGauge
	optionOperationsSampleFailure = failureCounter
	operationsMetricsState.lastAttempt = 0
	operationsMetricsState.counts = map[operationsCountKey]struct{}{
		{tenantID: "9", category: "asset_failed"}: {},
	}
	queryOptionOperationsMetrics = func(
		context.Context, sqlx.SqlConn, int64, int64, int64,
	) ([]*models.OptionOperationsMetric, []*models.OptionOperationsAmountMetric, error) {
		return nil, nil, errors.New("sample failed")
	}
	t.Cleanup(func() {
		optionOperationsSampleSuccess = originalSuccess
		optionOperationsSampleFailure = originalFailure
		queryOptionOperationsMetrics = originalQuery
		operationsMetricsState.lastAttempt = 0
		operationsMetricsState.counts = make(map[operationsCountKey]struct{})
		operationsMetricsState.amounts = make(map[operationsAmountKey]struct{})
	})

	if err := SampleOptionOperationsMetrics(context.Background(), nil, 2000); err == nil {
		t.Fatal("sample error must be returned to the caller for logging")
	}
	assertMetricCall(t, successGauge.calls, 0, 0)
	if len(failureCounter.calls) != 1 ||
		len(failureCounter.calls[0].labels) != 1 ||
		failureCounter.calls[0].labels[0] != "query" {
		t.Fatalf("failure calls=%+v", failureCounter.calls)
	}
	if len(operationsMetricsState.counts) != 1 {
		t.Fatal("failed sample must retain the last good series for alert continuity")
	}
}

func assertMetricCall(
	t *testing.T,
	calls []metricCall,
	index int,
	value float64,
	labels ...string,
) {
	t.Helper()
	if index >= len(calls) {
		t.Fatalf("missing metric call %d in %+v", index, calls)
	}
	call := calls[index]
	if call.value != value {
		t.Fatalf("metric call %d value=%v want=%v", index, call.value, value)
	}
	if len(call.labels) != len(labels) {
		t.Fatalf("metric call %d labels=%v want=%v", index, call.labels, labels)
	}
	for i := range labels {
		if call.labels[i] != labels[i] {
			t.Fatalf("metric call %d labels=%v want=%v", index, call.labels, labels)
		}
	}
}
