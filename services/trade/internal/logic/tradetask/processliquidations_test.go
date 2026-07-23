package tradetasklogic

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
