package middleware

import "testing"

func TestGetRequiredPermissionUsesDynamicRules(t *testing.T) {
	rules := []PermissionRule{
		mustRule(t, "GET", "/liquidity/admin/providers", "liquidity:provider:list"),
		mustRule(t, "POST", "/liquidity/admin/providers/{id}/test", "liquidity:provider:test"),
		mustRule(t, "POST", "/liquidity/admin/symbol-configs/{id}/start", "liquidity:strategy:start"),
	}
	tests := []struct {
		method string
		path   string
		want   string
	}{
		{"GET", "/liquidity/admin/providers", "liquidity:provider:list"},
		{"POST", "/liquidity/admin/providers/9/test", "liquidity:provider:test"},
		{"POST", "/liquidity/admin/symbol-configs/3/start", "liquidity:strategy:start"},
		{"DELETE", "/liquidity/admin/providers/9", ""},
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
