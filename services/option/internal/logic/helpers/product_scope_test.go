package helpers

import (
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/config"
	"wklive/services/option/models"
)

func TestContractLaunchProductScopeReadyFailClosed(t *testing.T) {
	contract := &models.TOptionContract{
		ExerciseStyle:    int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		SettlementType:   int64(option.SettlementType_SETTLEMENT_TYPE_CASH),
		SellerMarginMode: int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED),
	}
	if ok, reason := ContractLaunchProductScopeReady(contract, config.ProductScope{}); ok || reason != ProductScopeSellerTradingDisabled {
		t.Fatalf("expected seller gate, ok=%v reason=%s", ok, reason)
	}
	scope := config.ProductScope{SellerTradingEnabled: true}
	if ok, reason := ContractLaunchProductScopeReady(contract, scope); !ok {
		t.Fatalf("expected isolated seller contract enabled, reason=%s", reason)
	}
	contract.SellerMarginMode = int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO)
	if ok, reason := ContractLaunchProductScopeReady(contract, scope); ok || reason != ProductScopePortfolioMarginDisabled {
		t.Fatalf("expected portfolio gate, ok=%v reason=%s", ok, reason)
	}
	contract.SettlementType = int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL)
	if ok, reason := ContractLaunchProductScopeReady(contract, scope); ok || reason != ProductScopePhysicalDeliveryDisabled {
		t.Fatalf("expected physical gate first, ok=%v reason=%s", ok, reason)
	}
}

func TestOrderProductScopeReadyKeepsCloseAvailable(t *testing.T) {
	contract := &models.TOptionContract{
		SettlementType:   int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL),
		SellerMarginMode: int64(option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY),
	}
	if ok, reason := OrderProductScopeReady(contract, config.ProductScope{}, common.Side_SIDE_SELL, option.PositionEffect_POSITION_EFFECT_OPEN, common.YesNo_YES_NO_NO); ok || reason != ProductScopePhysicalDeliveryDisabled {
		t.Fatalf("expected physical open gate, ok=%v reason=%s", ok, reason)
	}
	if ok, reason := OrderProductScopeReady(contract, config.ProductScope{}, common.Side_SIDE_SELL, option.PositionEffect_POSITION_EFFECT_CLOSE, common.YesNo_YES_NO_NO); !ok {
		t.Fatalf("close must remain available, reason=%s", reason)
	}
	if ok, reason := OrderProductScopeReady(contract, config.ProductScope{}, common.Side_SIDE_BUY, option.PositionEffect_POSITION_EFFECT_CLOSE, common.YesNo_YES_NO_YES); ok || reason != ProductScopeMMPDisabled {
		t.Fatalf("MMP gate must apply to close orders, ok=%v reason=%s", ok, reason)
	}
}

func TestOrderProductScopeReadyBlocksBothSidesOfNewSellerRisk(t *testing.T) {
	contract := &models.TOptionContract{
		SettlementType:   int64(option.SettlementType_SETTLEMENT_TYPE_CASH),
		SellerMarginMode: int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED),
	}
	for _, side := range []common.Side{common.Side_SIDE_BUY, common.Side_SIDE_SELL} {
		if ok, reason := OrderProductScopeReady(
			contract, config.ProductScope{}, side,
			option.PositionEffect_POSITION_EFFECT_OPEN, common.YesNo_YES_NO_NO,
		); ok || reason != ProductScopeSellerTradingDisabled {
			t.Fatalf("expected seller gate for side=%s, ok=%v reason=%s", side, ok, reason)
		}
	}
}
