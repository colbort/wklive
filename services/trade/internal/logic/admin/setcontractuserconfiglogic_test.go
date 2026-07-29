package adminlogic

import (
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"
)

func TestValidateContractUserConfigKeepsCrossGateClosed(t *testing.T) {
	req := &trade.SetContractUserConfigReq{
		TenantId: 1, UserId: 2, SymbolId: 3,
		PositionMode:    trade.PositionMode_POSITION_MODE_ONE_WAY,
		MarginMode:      trade.MarginMode_MARGIN_MODE_CROSS,
		DefaultLeverage: 10,
	}
	if err := validateContractUserConfigInput(req, &models.TTradeSymbolContract{SupportCross: 1}, false); err == nil {
		t.Fatal("cross configuration must remain closed before account liquidation is enabled")
	}
	if err := validateContractUserConfigInput(req, &models.TTradeSymbolContract{SupportCross: 1}, true); err != nil {
		t.Fatalf("validated cross configuration should pass once the gate is enabled: %v", err)
	}
}

func TestContractUserModeChanged(t *testing.T) {
	current := &models.TContractUserConfig{
		PositionMode: int64(trade.PositionMode_POSITION_MODE_ONE_WAY),
		MarginMode:   int64(trade.MarginMode_MARGIN_MODE_ISOLATED),
	}
	req := &trade.SetContractUserConfigReq{
		PositionMode: trade.PositionMode_POSITION_MODE_ONE_WAY,
		MarginMode:   trade.MarginMode_MARGIN_MODE_ISOLATED,
	}
	if contractUserModeChanged(current, req) {
		t.Fatal("identical modes must not require a flat risk unit")
	}
	req.PositionMode = trade.PositionMode_POSITION_MODE_HEDGE
	if !contractUserModeChanged(current, req) {
		t.Fatal("position mode change must require a flat risk unit")
	}
}
