package tasklogic

import (
	"testing"

	"wklive/proto/option"
	"wklive/services/option/models"
)

func TestContractRequestsPlatformBackstop(t *testing.T) {
	if contractRequestsPlatformBackstop(nil) {
		t.Fatal("nil contract must not request platform backstop")
	}
	contract := &models.TOptionContract{
		LiquidationDeficitPolicy: int64(
			option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW,
		),
	}
	if contractRequestsPlatformBackstop(contract) {
		t.Fatal("manual-review policy must not request platform backstop")
	}
	contract.LiquidationDeficitPolicy = int64(
		option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_PLATFORM_BACKSTOP,
	)
	if !contractRequestsPlatformBackstop(contract) {
		t.Fatal("platform-backstop policy must be detected")
	}
	if platformBackstopRuntimeEnabled(contract, false) {
		t.Fatal("platform backstop must fail closed when the runtime gate is disabled")
	}
	if !platformBackstopRuntimeEnabled(contract, true) {
		t.Fatal("platform backstop should be available only when policy and gate are both enabled")
	}
	contract.LiquidationDeficitPolicy = int64(
		option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW,
	)
	if platformBackstopRuntimeEnabled(contract, true) {
		t.Fatal("runtime gate must not override a manual-review contract policy")
	}
}
