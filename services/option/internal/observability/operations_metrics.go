package observability

import (
	"context"
	"strconv"
	"sync"

	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/metric"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const operationsSampleIntervalSeconds int64 = 15

var (
	optionOperationsCount gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "operations_count",
		Help:      "Current Option operational exception or backlog count by tenant and category.",
		Labels:    []string{"tenant_id", "category"},
	})
	optionOperationsOldestTimestamp gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "operations_oldest_timestamp_seconds",
		Help:      "Oldest source timestamp or earliest relevant deadline for the current Option operational condition.",
		Labels:    []string{"tenant_id", "category"},
	})
	optionOperationsAmount gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "operations_amount",
		Help:      "Current Option financial amount or insurance takeover exposure by tenant, category, and currency or underlying coin.",
		Labels:    []string{"tenant_id", "category", "coin"},
	})
	optionOperationsSampleSuccess gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "operations_sample_success",
		Help:      "Whether the latest Option operations metrics sample succeeded.",
	})
	optionOperationsLastSuccessTimestamp gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "operations_last_success_timestamp_seconds",
		Help:      "Unix timestamp of the latest successful Option operations metrics sample.",
	})
	optionOperationsSampleFailure counterVec = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "operations_sample_failure_total",
		Help:      "Failed attempts to sample Option operational metrics.",
		Labels:    []string{"stage"},
	})
	queryOptionOperationsMetrics = models.QueryOptionOperationsMetrics
	operationsMetricsState       = struct {
		sync.Mutex
		lastAttempt int64
		counts      map[operationsCountKey]struct{}
		amounts     map[operationsAmountKey]struct{}
	}{
		counts:  make(map[operationsCountKey]struct{}),
		amounts: make(map[operationsAmountKey]struct{}),
	}
)

type operationsCountKey struct {
	tenantID string
	category string
}

type operationsAmountKey struct {
	tenantID string
	category string
	coin     string
}

func SampleOptionOperationsMetrics(
	ctx context.Context,
	conn sqlx.SqlConn,
	now int64,
) error {
	operationsMetricsState.Lock()
	defer operationsMetricsState.Unlock()
	if operationsMetricsState.lastAttempt > 0 &&
		now-operationsMetricsState.lastAttempt < operationsSampleIntervalSeconds {
		return nil
	}
	operationsMetricsState.lastAttempt = now
	counts, amounts, err := queryOptionOperationsMetrics(ctx, conn, now-60, now-60, now)
	if err != nil {
		optionOperationsSampleSuccess.Set(0)
		optionOperationsSampleFailure.Inc("query")
		return err
	}

	for key := range operationsMetricsState.counts {
		optionOperationsCount.Set(0, key.tenantID, key.category)
		optionOperationsOldestTimestamp.Set(0, key.tenantID, key.category)
	}
	for key := range operationsMetricsState.amounts {
		optionOperationsAmount.Set(0, key.tenantID, key.category, key.coin)
	}
	nextCounts := make(map[operationsCountKey]struct{}, len(counts))
	for _, item := range counts {
		if item == nil {
			continue
		}
		tenantID := strconv.FormatInt(item.TenantID, 10)
		key := operationsCountKey{tenantID: tenantID, category: item.Category}
		optionOperationsCount.Set(float64(item.Count), tenantID, item.Category)
		optionOperationsOldestTimestamp.Set(float64(item.Oldest), tenantID, item.Category)
		nextCounts[key] = struct{}{}
	}
	nextAmounts := make(map[operationsAmountKey]struct{}, len(amounts))
	for _, item := range amounts {
		if item == nil {
			continue
		}
		tenantID := strconv.FormatInt(item.TenantID, 10)
		key := operationsAmountKey{
			tenantID: tenantID,
			category: item.Category,
			coin:     item.Coin,
		}
		optionOperationsAmount.Set(
			item.Amount.InexactFloat64(), tenantID, item.Category, item.Coin,
		)
		nextAmounts[key] = struct{}{}
	}
	operationsMetricsState.counts = nextCounts
	operationsMetricsState.amounts = nextAmounts
	optionOperationsSampleSuccess.Set(1)
	optionOperationsLastSuccessTimestamp.Set(float64(now))
	return nil
}
