package tasklogic

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestBuildReconciliationRecord(t *testing.T) {
	key := reconciliationKey{TenantId: 1, CoinSymbol: "USDT"}
	matched := buildReconciliationRecord(key, 20260803, 1, reconciliationAmounts{
		ActivePrincipal: decimal.NewFromInt(10), ProductStaked: decimal.NewFromInt(10),
		PositionStaked: decimal.NewFromInt(10), AssetLocked: decimal.NewFromInt(10),
		RewardLogAmount: decimal.RequireFromString("1.25"), RewardPlatformAmount: decimal.RequireFromString("1.25"),
		FeeLogAmount: decimal.RequireFromString("0.5"), FeePlatformAmount: decimal.RequireFromString("0.5"),
	})
	if matched.Status != stakeReconciliationStatusMatched || matched.Detail != "" {
		t.Fatalf("expected matched reconciliation, got status=%d detail=%q", matched.Status, matched.Detail)
	}

	diff := buildReconciliationRecord(key, 20260803, 1, reconciliationAmounts{
		ActivePrincipal: decimal.NewFromInt(10), ProductStaked: decimal.NewFromInt(11),
		PositionStaked: decimal.NewFromInt(10), AssetLocked: decimal.NewFromInt(9),
		RewardLogAmount: decimal.NewFromInt(1), RewardPlatformAmount: decimal.NewFromInt(2),
	})
	if diff.Status != stakeReconciliationStatusDiff || diff.ProductDiff.String() != "1" || diff.LockDiff.String() != "-1" || diff.RewardDiff.String() != "1" {
		t.Fatalf("unexpected diff reconciliation: %+v", diff)
	}
}

func TestUTCDate(t *testing.T) {
	millis := time.Date(2026, 8, 3, 23, 59, 0, 0, time.FixedZone("HKT", 8*60*60)).UnixMilli()
	if got := utcDate(millis); got != 20260803 {
		t.Fatalf("expected UTC date 20260803, got %d", got)
	}
}
