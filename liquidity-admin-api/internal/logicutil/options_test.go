package logicutil

import "testing"

func TestLiquidityOptionsContainsRequiredGroups(t *testing.T) {
	required := map[string]bool{
		"providerType":          false,
		"providerEnvironment":   false,
		"liquidityMode":         false,
		"symbolLiquidityStatus": false,
		"quoteOrderStatus":      false,
		"externalOrderStatus":   false,
		"hedgeStatus":           false,
		"riskEventStatus":       false,
		"reconcileStatus":       false,
	}
	for _, group := range LiquidityOptions() {
		if _, ok := required[group.Key]; ok {
			required[group.Key] = len(group.Options) > 0
		}
	}
	for key, found := range required {
		if !found {
			t.Errorf("missing or empty options group %q", key)
		}
	}
}
