package middleware

import "testing"

func TestGetRequiredPermissionUsesDynamicRules(t *testing.T) {
	rules := []PermissionRule{
		mustRule(t, "GET", "/admin/liquidity/providers", "liquidity:provider:list"),
		mustRule(t, "GET", "/admin/liquidity/config-options", "liquidity:strategy:list"),
		mustRule(t, "POST", "/admin/liquidity/providers/provision", "liquidity:provider:add"),
		mustRule(t, "POST", "/admin/liquidity/providers/{id}/test", "liquidity:provider:test"),
		mustRule(t, "GET", "/admin/liquidity/symbol-configs/{id}", "liquidity:strategy:detail"),
		mustRule(t, "PUT", "/admin/liquidity/symbol-configs/{id}", "liquidity:strategy:update"),
		mustRule(t, "POST", "/admin/liquidity/symbol-configs/{id}/start", "liquidity:strategy:start"),
		mustRule(t, "GET", "/admin/liquidity/providers/{id}", "liquidity:provider:detail"),
		mustRule(t, "PUT", "/admin/liquidity/providers/{id}", "liquidity:provider:update"),
		mustRule(t, "POST", "/admin/liquidity/symbol-configs/{id}/cancel-quotes", "liquidity:quote:cancel-all"),
		mustRule(t, "GET", "/admin/liquidity/quote-cycles", "liquidity:quote-cycle:list"),
		mustRule(t, "GET", "/admin/liquidity/external-fills", "liquidity:external-fill:list"),
		mustRule(t, "POST", "/admin/liquidity/external-orders/{id}/cancel", "liquidity:external-order:cancel"),
		mustRule(t, "POST", "/admin/liquidity/hedge-tasks/manual", "liquidity:hedge:create"),
		mustRule(t, "POST", "/admin/liquidity/hedge-tasks/{id}/cancel", "liquidity:hedge:cancel"),
		mustRule(t, "POST", "/admin/liquidity/hedge-tasks/{id}/retry", "liquidity:hedge:retry"),
		mustRule(t, "GET", "/admin/liquidity/inventory-snapshots", "liquidity:inventory:list"),
		mustRule(t, "GET", "/admin/liquidity/inventory-snapshots/latest", "liquidity:inventory:latest"),
		mustRule(t, "POST", "/admin/liquidity/risk-events/{id}/resolve", "liquidity:risk:resolve"),
		mustRule(t, "POST", "/admin/liquidity/reconcile-batches/run", "liquidity:reconcile:run"),
		mustRule(t, "GET", "/admin/liquidity/reconcile-batches/{batchId}/details", "liquidity:reconcile:detail"),
		mustRule(t, "POST", "/admin/liquidity/reconcile-differences/{id}/resolve", "liquidity:reconcile:resolve"),
	}
	tests := []struct {
		method string
		path   string
		want   string
	}{
		{"GET", "/admin/liquidity/providers", "liquidity:provider:list"},
		{"GET", "/admin/liquidity/config-options", "liquidity:strategy:list"},
		{"POST", "/admin/liquidity/providers/provision", "liquidity:provider:add"},
		{"POST", "/admin/liquidity/providers/9/test", "liquidity:provider:test"},
		{"GET", "/admin/liquidity/symbol-configs/3", "liquidity:strategy:detail"},
		{"PUT", "/admin/liquidity/symbol-configs/3", "liquidity:strategy:update"},
		{"POST", "/admin/liquidity/symbol-configs/3/start", "liquidity:strategy:start"},
		{"GET", "/admin/liquidity/providers/9", "liquidity:provider:detail"},
		{"PUT", "/admin/liquidity/providers/9", "liquidity:provider:update"},
		{"POST", "/admin/liquidity/symbol-configs/3/cancel-quotes", "liquidity:quote:cancel-all"},
		{"GET", "/admin/liquidity/quote-cycles", "liquidity:quote-cycle:list"},
		{"GET", "/admin/liquidity/external-fills", "liquidity:external-fill:list"},
		{"POST", "/admin/liquidity/external-orders/5/cancel", "liquidity:external-order:cancel"},
		{"POST", "/admin/liquidity/hedge-tasks/manual", "liquidity:hedge:create"},
		{"POST", "/admin/liquidity/hedge-tasks/7/cancel", "liquidity:hedge:cancel"},
		{"POST", "/admin/liquidity/hedge-tasks/7/retry", "liquidity:hedge:retry"},
		{"GET", "/admin/liquidity/inventory-snapshots", "liquidity:inventory:list"},
		{"GET", "/admin/liquidity/inventory-snapshots/latest", "liquidity:inventory:latest"},
		{"POST", "/admin/liquidity/risk-events/8/resolve", "liquidity:risk:resolve"},
		{"POST", "/admin/liquidity/reconcile-batches/run", "liquidity:reconcile:run"},
		{"GET", "/admin/liquidity/reconcile-batches/10/details", "liquidity:reconcile:detail"},
		{"POST", "/admin/liquidity/reconcile-differences/11/resolve", "liquidity:reconcile:resolve"},
		{"DELETE", "/admin/liquidity/providers/9", ""},
	}
	for _, test := range tests {
		if got := getRequiredPermission(rules, test.path, test.method); got != test.want {
			t.Errorf("%s %s: got %q, want %q", test.method, test.path, got, test.want)
		}
	}
}

func mustRule(t *testing.T, method, path, permission string) PermissionRule {
	t.Helper()
	pattern, staticSegs, err := compilePathPattern(path)
	if err != nil {
		t.Fatal(err)
	}
	return PermissionRule{
		Method: method, Path: normalizePath(path), PermKey: permission,
		Pattern: pattern, StaticSegs: staticSegs,
	}
}
