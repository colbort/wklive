package tasklogic

import (
	"testing"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"

	"github.com/shopspring/decimal"
)

func TestADLTakeoverQtyCapsAtBankruptRemainingQty(t *testing.T) {
	got := adlTakeoverQty(decimal.NewFromInt(100), decimal.NewFromInt(3), decimal.NewFromInt(50), decimal.NewFromInt(2))
	if !got.Equal(decimal.NewFromInt(3)) {
		t.Fatalf("takeover qty = %s, want 3", got)
	}
}

func TestValidateLiquidationAssetResponseRejectsEmptyResponse(t *testing.T) {
	if err := validateLiquidationAssetResponse(nil); err == nil {
		t.Fatal("nil Asset response must be rejected")
	}
	if err := validateLiquidationAssetResponse(&asset.ChangeAssetResp{}); err == nil {
		t.Fatal("Asset response without base must be rejected")
	}
}

func TestValidateLiquidationAssetResponse(t *testing.T) {
	if err := validateLiquidationAssetResponse(&asset.ChangeAssetResp{Base: &common.RespBase{Code: 500, Msg: "rejected"}}); err == nil {
		t.Fatal("non-success Asset response must be rejected")
	}
	if err := validateLiquidationAssetResponse(&asset.ChangeAssetResp{Base: &common.RespBase{Code: 200}}); err != nil {
		t.Fatalf("successful Asset response rejected: %v", err)
	}
}

func TestLiquidationStageRankDoesNotRegress(t *testing.T) {
	stages := []trade.LiquidationStatus{
		trade.LiquidationStatus_LIQUIDATION_STATUS_PENDING_TAKEOVER,
		trade.LiquidationStatus_LIQUIDATION_STATUS_LIQUIDATING,
		trade.LiquidationStatus_LIQUIDATION_STATUS_INSURANCE_FUND,
		trade.LiquidationStatus_LIQUIDATION_STATUS_ADL,
		trade.LiquidationStatus_LIQUIDATION_STATUS_COMPLETED,
		trade.LiquidationStatus_LIQUIDATION_STATUS_MANUAL_REVIEW,
	}
	for i := 1; i < len(stages); i++ {
		if liquidationStageRank(stages[i]) <= liquidationStageRank(stages[i-1]) {
			t.Fatalf("stage rank did not advance: %v -> %v", stages[i-1], stages[i])
		}
	}
}

func TestADLTakeoverQtyRejectsInvalidInputs(t *testing.T) {
	if got := adlTakeoverQty(decimal.NewFromInt(10), decimal.NewFromInt(10), decimal.NewFromInt(5), decimal.Zero); !got.IsZero() {
		t.Fatalf("takeover qty = %s, want 0", got)
	}
}

func TestADLMarginReleaseKeepsMarginBucketsSeparate(t *testing.T) {
	position, isolated := adlMarginRelease(decimal.NewFromInt(80), decimal.NewFromInt(20), decimal.NewFromInt(2), decimal.NewFromInt(10))
	if !position.Equal(decimal.NewFromInt(16)) || !isolated.Equal(decimal.NewFromInt(4)) {
		t.Fatalf("released position=%s isolated=%s, want 16 and 4", position, isolated)
	}
}

func TestAutomaticLiquidationProductionGate(t *testing.T) {
	pending := int64(trade.LiquidationStatus_LIQUIDATION_STATUS_PENDING_TAKEOVER)
	if !shouldHoldLiquidationForManual(false, pending) {
		t.Fatal("disabled automatic liquidation must hold a new takeover for manual review")
	}
	if shouldHoldLiquidationForManual(true, pending) {
		t.Fatal("accepted and explicitly enabled automatic liquidation was held")
	}
	if shouldHoldLiquidationForManual(false, int64(trade.LiquidationStatus_LIQUIDATION_STATUS_INSURANCE_FUND)) {
		t.Fatal("an already-started money saga must remain recoverable after the gate is disabled")
	}
}

func TestSplitLiquidationEquity(t *testing.T) {
	tests := []struct {
		name         string
		equity       string
		nominalFee   string
		wantFee      string
		wantResidual string
		wantDeficit  string
	}{
		{name: "fee and residual", equity: "20", nominalFee: "3", wantFee: "3", wantResidual: "17", wantDeficit: "0"},
		{name: "fee capped by equity", equity: "2", nominalFee: "3", wantFee: "2", wantResidual: "0", wantDeficit: "0"},
		{name: "bankruptcy excludes fee", equity: "-5", nominalFee: "3", wantFee: "0", wantResidual: "0", wantDeficit: "5"},
		{name: "negative fee rejected", equity: "5", nominalFee: "-1", wantFee: "0", wantResidual: "5", wantDeficit: "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fee, residual, deficit := splitLiquidationEquity(
				decimal.RequireFromString(tt.equity),
				decimal.RequireFromString(tt.nominalFee),
			)
			if fee.String() != tt.wantFee || residual.String() != tt.wantResidual || deficit.String() != tt.wantDeficit {
				t.Fatalf("got fee=%s residual=%s deficit=%s, want %s/%s/%s", fee, residual, deficit, tt.wantFee, tt.wantResidual, tt.wantDeficit)
			}
		})
	}
}
