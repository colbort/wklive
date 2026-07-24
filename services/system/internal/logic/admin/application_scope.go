package adminlogic

import "wklive/proto/system"

func normalizeApplicationScope(scope system.ApplicationScope) system.ApplicationScope {
	if scope == system.ApplicationScope_APPLICATION_SCOPE_LIQUIDITY {
		return scope
	}
	return system.ApplicationScope_APPLICATION_SCOPE_ADMIN
}

func validApplicationScope(scope system.ApplicationScope) bool {
	return scope == system.ApplicationScope_APPLICATION_SCOPE_ADMIN ||
		scope == system.ApplicationScope_APPLICATION_SCOPE_LIQUIDITY
}
