package observability

import (
	"strconv"
	"sync"

	"github.com/zeromicro/go-zero/core/metric"
)

type RiskScanTenantResult struct {
	TenantID     int64
	TotalGroups  int64
	FailedGroups int64
}

var (
	optionRiskScanGroups gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "risk_scan_groups",
		Help:      "Wallet risk groups evaluated in the latest completed Option risk scan.",
		Labels:    []string{"tenant_id"},
	})
	optionRiskScanFailedGroups gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "risk_scan_failed_groups",
		Help:      "Wallet risk groups that failed in the latest completed Option risk scan.",
		Labels:    []string{"tenant_id"},
	})
	optionRiskScanFailureRatio gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "risk_scan_failure_ratio",
		Help:      "Failed wallet risk groups divided by all groups in the latest completed scan.",
		Labels:    []string{"tenant_id"},
	})
	optionRiskScanLastCompletedTimestamp gaugeVec = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "risk_scan_last_completed_timestamp_seconds",
		Help:      "Unix timestamp of the latest completed Option risk scan by tenant.",
		Labels:    []string{"tenant_id"},
	})
	optionRiskScanExecutionFailure counterVec = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: "wklive",
		Subsystem: "option",
		Name:      "risk_scan_execution_failure_total",
		Help:      "Option risk scan failures before or after per-wallet evaluation.",
		Labels:    []string{"tenant_scope", "stage"},
	})
	riskScanMetricState = struct {
		sync.Mutex
		tenants map[string]struct{}
	}{
		tenants: make(map[string]struct{}),
	}
)

func PublishRiskScanResults(results []RiskScanTenantResult, completedAt, tenantScopeID int64) {
	riskScanMetricState.Lock()
	defer riskScanMetricState.Unlock()
	published := make(map[string]struct{}, len(results))
	for _, result := range results {
		if result.TenantID <= 0 {
			continue
		}
		tenantID := strconv.FormatInt(result.TenantID, 10)
		failed := result.FailedGroups
		if failed < 0 {
			failed = 0
		}
		total := result.TotalGroups
		if total < failed {
			total = failed
		}
		ratio := float64(0)
		if total > 0 {
			ratio = float64(failed) / float64(total)
		}
		optionRiskScanGroups.Set(float64(total), tenantID)
		optionRiskScanFailedGroups.Set(float64(failed), tenantID)
		optionRiskScanFailureRatio.Set(ratio, tenantID)
		optionRiskScanLastCompletedTimestamp.Set(float64(completedAt), tenantID)
		published[tenantID] = struct{}{}
		riskScanMetricState.tenants[tenantID] = struct{}{}
	}
	if tenantScopeID > 0 {
		// A tenant-scoped replay updates only that tenant. The caller supplies an
		// explicit zero result when it has no active groups.
		return
	}
	// tenantScopeID=0 is the production full-tenant scan and is authoritative
	// for disappearance: clear series not present in this completed global run.
	for tenantID := range riskScanMetricState.tenants {
		if _, ok := published[tenantID]; ok {
			continue
		}
		optionRiskScanGroups.Set(0, tenantID)
		optionRiskScanFailedGroups.Set(0, tenantID)
		optionRiskScanFailureRatio.Set(0, tenantID)
		optionRiskScanLastCompletedTimestamp.Set(float64(completedAt), tenantID)
		delete(riskScanMetricState.tenants, tenantID)
	}
}

func RecordRiskScanExecutionFailure(tenantID int64, stage string) {
	optionRiskScanExecutionFailure.Inc(tenantScope(tenantID), stage)
}
