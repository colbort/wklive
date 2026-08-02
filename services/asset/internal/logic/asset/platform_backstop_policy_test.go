package assetlogic

import (
	"testing"
	"time"

	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func effectiveBackstopPolicy(mode int64) *models.TAssetBackstopPolicy {
	return &models.TAssetBackstopPolicy{
		Status: backstopPolicyStatusApproved, Mode: mode,
		PerRequestLimit: decimal.NewFromInt(10), DailyLimit: decimal.NewFromInt(100),
		BalanceFloor: decimal.Zero, EffectiveFrom: 1000, EffectiveUntil: 3000,
	}
}

func TestPlatformBackstopUsageDayUsesUTC(t *testing.T) {
	before := time.Date(2026, 8, 2, 23, 59, 59, 999_000_000, time.UTC).UnixMilli()
	if got := platformBackstopUsageDay(before); got != "20260802" {
		t.Fatalf("before UTC midnight day=%s", got)
	}
	if got := platformBackstopUsageDay(before + 1); got != "20260803" {
		t.Fatalf("at UTC midnight day=%s", got)
	}
	plusEight := time.FixedZone("UTC+8", 8*60*60)
	local := time.Date(2026, 8, 3, 7, 59, 59, 999_000_000, plusEight).UnixMilli()
	if got := platformBackstopUsageDay(local); got != "20260802" {
		t.Fatalf("local timezone changed UTC bucket day=%s", got)
	}
}

func TestValidateEffectiveBackstopPolicy(t *testing.T) {
	prefunded := effectiveBackstopPolicy(backstopPolicyModePrefunded)
	if err := validateEffectiveBackstopPolicy(prefunded, decimal.NewFromInt(10), 2000); err != nil {
		t.Fatalf("exact prefunded request limit rejected: %v", err)
	}
	credit := effectiveBackstopPolicy(backstopPolicyModeCreditFloor)
	credit.BalanceFloor = decimal.NewFromInt(-50)
	if err := validateEffectiveBackstopPolicy(credit, decimal.NewFromInt(10), 2000); err != nil {
		t.Fatalf("valid credit-floor policy rejected: %v", err)
	}
	tests := []struct {
		name   string
		policy *models.TAssetBackstopPolicy
		amount decimal.Decimal
		now    int64
	}{
		{name: "missing policy", policy: nil, amount: decimal.NewFromInt(1), now: 2000},
		{name: "disabled", policy: effectiveBackstopPolicy(backstopPolicyModeDisabled), amount: decimal.NewFromInt(1), now: 2000},
		{name: "draft", policy: func() *models.TAssetBackstopPolicy {
			p := effectiveBackstopPolicy(backstopPolicyModePrefunded)
			p.Status = 1
			return p
		}(), amount: decimal.NewFromInt(1), now: 2000},
		{name: "not effective", policy: effectiveBackstopPolicy(backstopPolicyModePrefunded), amount: decimal.NewFromInt(1), now: 999},
		{name: "expired", policy: effectiveBackstopPolicy(backstopPolicyModePrefunded), amount: decimal.NewFromInt(1), now: 3000},
		{name: "over request limit", policy: effectiveBackstopPolicy(backstopPolicyModePrefunded), amount: decimal.NewFromInt(11), now: 2000},
		{name: "request limit above daily", policy: func() *models.TAssetBackstopPolicy {
			p := effectiveBackstopPolicy(backstopPolicyModePrefunded)
			p.PerRequestLimit = decimal.NewFromInt(101)
			return p
		}(), amount: decimal.NewFromInt(1), now: 2000},
		{name: "prefunded negative floor", policy: func() *models.TAssetBackstopPolicy {
			p := effectiveBackstopPolicy(backstopPolicyModePrefunded)
			p.BalanceFloor = decimal.NewFromInt(-1)
			return p
		}(), amount: decimal.NewFromInt(1), now: 2000},
		{name: "credit zero floor", policy: effectiveBackstopPolicy(backstopPolicyModeCreditFloor), amount: decimal.NewFromInt(1), now: 2000},
		{name: "unknown mode", policy: effectiveBackstopPolicy(99), amount: decimal.NewFromInt(1), now: 2000},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEffectiveBackstopPolicy(tc.policy, tc.amount, tc.now)
			if err == nil {
				t.Fatalf("invalid effective policy accepted: %+v", tc.policy)
			}
			if status.Code(err) != codes.FailedPrecondition {
				t.Fatalf("business rejection code=%s want=%s err=%v", status.Code(err), codes.FailedPrecondition, err)
			}
		})
	}
}
