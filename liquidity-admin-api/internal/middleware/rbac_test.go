package middleware

import "testing"

func TestGetRequiredPermissionUsesDynamicRules(t *testing.T) {
	rules := []PermissionRule{
		mustRule(t, "GET", "/admin/liquidity/providers", "liquidity:provider:list"),
		mustRule(t, "GET", "/admin/liquidity/config-options", "liquidity:strategy:list"),
		mustRule(t, "POST", "/admin/liquidity/providers/provision", "liquidity:provider:add"),
		mustRule(t, "POST", "/admin/liquidity/providers/{id}/test", "liquidity:provider:test"),
		mustRule(t, "POST", "/admin/liquidity/symbol-configs/{id}/start", "liquidity:strategy:start"),
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
		{"POST", "/admin/liquidity/symbol-configs/3/start", "liquidity:strategy:start"},
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
