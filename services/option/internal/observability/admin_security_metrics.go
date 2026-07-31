package observability

import "github.com/zeromicro/go-zero/core/metric"

var adminRejectedMutation counterVec = metric.NewCounterVec(&metric.CounterVecOpts{
	Namespace: "wklive",
	Subsystem: "option",
	Name:      "admin_rejected_mutation_total",
	Help:      "Rejected Option administrator attempts to mutate governed state or economic fields.",
	Labels:    []string{"tenant_scope", "object_type", "reason"},
})

func RecordAdminRejectedMutation(tenantID int64, objectType, reason string) {
	adminRejectedMutation.Inc(
		tenantScope(tenantID),
		boundedAdminObjectType(objectType),
		boundedAdminMutationReason(reason),
	)
}

func boundedAdminObjectType(value string) string {
	switch value {
	case "contract":
		return value
	default:
		return "other"
	}
}

func boundedAdminMutationReason(value string) string {
	switch value {
	case "status_bypass", "series_economics", "listed_economics":
		return value
	default:
		return "other"
	}
}
