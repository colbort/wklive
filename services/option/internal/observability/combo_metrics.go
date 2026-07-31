package observability

import (
	"strconv"

	"github.com/zeromicro/go-zero/core/metric"
)

type (
	counterVec interface {
		Inc(labels ...string)
	}
	gaugeVec interface {
		Set(value float64, labels ...string)
	}
)

var (
	comboIsolationViolation counterVec = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "combo_isolation_violation_total",
		Help:      "Combo shadow orders rejected from a simple-order path.",
		Labels:    []string{"tenant_scope", "path"},
	})
	comboDebitBarrierViolation counterVec = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "combo_debit_barrier_violation_total",
		Help:      "Combo position events rejected because the match group or debit barrier was incomplete.",
		Labels:    []string{"tenant_scope"},
	})
	comboDebitBarrierStaleEvents gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "combo_debit_barrier_stale_events",
		Help:      "Current combo position events older than 60 seconds blocked by an incomplete cross-leg debit barrier.",
		Labels:    []string{"tenant_scope"},
	})
	comboObservabilityQueryFailure counterVec = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "combo_observability_query_failure_total",
		Help:      "Failures while querying combo operational runtime watermarks.",
		Labels:    []string{"tenant_scope", "query"},
	})
)

func RecordComboIsolationViolation(tenantID int64, path string) {
	comboIsolationViolation.Inc(tenantScope(tenantID), path)
}

func RecordComboDebitBarrierViolation(tenantID int64) {
	comboDebitBarrierViolation.Inc(tenantScope(tenantID))
}

func SetComboDebitBarrierStaleEvents(tenantID, count int64) {
	comboDebitBarrierStaleEvents.Set(float64(count), tenantScope(tenantID))
}

func RecordComboObservabilityQueryFailure(tenantID int64, query string) {
	comboObservabilityQueryFailure.Inc(tenantScope(tenantID), query)
}

func tenantScope(tenantID int64) string {
	if tenantID <= 0 {
		return "all"
	}
	return strconv.FormatInt(tenantID, 10)
}
