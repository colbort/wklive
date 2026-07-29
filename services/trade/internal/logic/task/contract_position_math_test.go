package tasklogic

import (
	"testing"

	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func TestContractAveragePrice(t *testing.T) {
	linear := contractAveragePrice(decimal.NewFromInt(100), decimal.NewFromInt(2), decimal.NewFromInt(110), decimal.NewFromInt(2), int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR))
	if !linear.Equal(decimal.NewFromInt(105)) {
		t.Fatalf("linear average = %s, want 105", linear)
	}
	inverse := contractAveragePrice(decimal.NewFromInt(100), decimal.NewFromInt(1), decimal.NewFromInt(200), decimal.NewFromInt(1), int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE))
	if !inverse.Equal(decimal.NewFromInt(2).Div(decimal.RequireFromString("0.015"))) {
		t.Fatalf("inverse average = %s", inverse)
	}
}

func TestValidateContractPriceAndQuantityScale(t *testing.T) {
	symbol := &models.TTradeSymbol{PriceScale: 2, QtyScale: 3, PriceTick: decimal.RequireFromString("0.01"), QtyStep: decimal.RequireFromString("0.001")}
	if err := validateSymbolOrderIncrements(symbol, trade.OrderType_ORDER_TYPE_LIMIT, decimal.RequireFromString("10.12"), decimal.RequireFromString("1.234")); err != nil {
		t.Fatalf("valid price/qty rejected: %v", err)
	}
	if err := validateSymbolOrderIncrements(symbol, trade.OrderType_ORDER_TYPE_LIMIT, decimal.RequireFromString("10.123"), decimal.RequireFromString("1.234")); err == nil {
		t.Fatal("price exceeding configured scale should be rejected")
	}
	if err := validateSymbolOrderIncrements(symbol, trade.OrderType_ORDER_TYPE_LIMIT, decimal.RequireFromString("10.12"), decimal.RequireFromString("1.2345")); err == nil {
		t.Fatal("quantity exceeding configured scale should be rejected")
	}
}

func TestContractRealizedPnl(t *testing.T) {
	long := contractRealizedPnl(int64(trade.PositionSide_POSITION_SIDE_LONG), decimal.NewFromInt(100), decimal.NewFromInt(110), decimal.NewFromInt(2), decimal.NewFromInt(1), int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR))
	short := contractRealizedPnl(int64(trade.PositionSide_POSITION_SIDE_SHORT), decimal.NewFromInt(100), decimal.NewFromInt(90), decimal.NewFromInt(2), decimal.NewFromInt(1), int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR))
	if !long.Equal(decimal.NewFromInt(20)) || !short.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("linear pnl long=%s short=%s, want 20", long, short)
	}
	inverse := contractRealizedPnl(int64(trade.PositionSide_POSITION_SIDE_LONG), decimal.NewFromInt(100), decimal.NewFromInt(200), decimal.NewFromInt(100), decimal.NewFromInt(1), int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE))
	if !inverse.Equal(decimal.RequireFromString("0.5")) {
		t.Fatalf("inverse pnl = %s, want 0.5", inverse)
	}
}

func TestRecalculatePositionRisk(t *testing.T) {
	position := &models.TContractPosition{PositionSide: int64(trade.PositionSide_POSITION_SIDE_LONG), ContractValueType: int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR), Qty: decimal.NewFromInt(10), OpenAvgPrice: decimal.NewFromInt(100), MarkPrice: decimal.NewFromInt(110), PositionMargin: decimal.NewFromInt(100)}
	contract := &models.TTradeSymbolContract{ContractSize: decimal.NewFromInt(1), MaintenanceMarginRate: decimal.RequireFromString("0.01")}
	recalculatePositionRisk(position, contract)
	if !position.UnrealizedPnl.Equal(decimal.NewFromInt(100)) || !position.MaintenanceMargin.Equal(decimal.NewFromInt(11)) || !position.BankruptcyPrice.Equal(decimal.NewFromInt(90)) {
		t.Fatalf("unexpected risk projection: pnl=%s maintenance=%s bankruptcy=%s", position.UnrealizedPnl, position.MaintenanceMargin, position.BankruptcyPrice)
	}
	wantLiquidation := decimal.NewFromInt(90).Div(decimal.RequireFromString("0.99"))
	if !position.LiquidationPrice.Equal(wantLiquidation) {
		t.Fatalf("linear liquidation = %s, want %s", position.LiquidationPrice, wantLiquidation)
	}
}

func TestInverseMaintenanceUsesBaseSettlementUnit(t *testing.T) {
	position := &models.TContractPosition{PositionSide: int64(trade.PositionSide_POSITION_SIDE_LONG), ContractValueType: int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE), Qty: decimal.NewFromInt(100), OpenAvgPrice: decimal.NewFromInt(50000), MarkPrice: decimal.NewFromInt(50000), PositionMargin: decimal.RequireFromString("0.02")}
	contract := &models.TTradeSymbolContract{ContractSize: decimal.NewFromInt(100), MaintenanceMarginRate: decimal.RequireFromString("0.005")}
	recalculatePositionRisk(position, contract)
	if !position.MaintenanceMargin.Equal(decimal.RequireFromString("0.001")) {
		t.Fatalf("inverse maintenance = %s, want 0.001 base asset", position.MaintenanceMargin)
	}
	contracts := position.Qty.Mul(contract.ContractSize)
	equityAtLiquidation := position.PositionMargin.Add(contracts.Mul(decimal.NewFromInt(1).Div(position.OpenAvgPrice).Sub(decimal.NewFromInt(1).Div(position.LiquidationPrice))))
	maintenanceAtLiquidation := contracts.Div(position.LiquidationPrice).Mul(contract.MaintenanceMarginRate)
	if equityAtLiquidation.Sub(maintenanceAtLiquidation).Abs().GreaterThan(decimal.RequireFromString("0.000000000001")) {
		t.Fatalf("inverse liquidation equation mismatch: equity=%s maintenance=%s price=%s", equityAtLiquidation, maintenanceAtLiquidation, position.LiquidationPrice)
	}
}

func TestRiskTierMaintenanceAmountAndCrossBoundary(t *testing.T) {
	position := &models.TContractPosition{PositionSide: int64(trade.PositionSide_POSITION_SIDE_LONG), ContractValueType: int64(trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR), MarginMode: int64(trade.MarginMode_MARGIN_MODE_ISOLATED), Qty: decimal.NewFromInt(10), OpenAvgPrice: decimal.NewFromInt(100), MarkPrice: decimal.NewFromInt(100), PositionMargin: decimal.NewFromInt(100)}
	contract := &models.TTradeSymbolContract{ContractSize: decimal.NewFromInt(1), MaintenanceMarginRate: decimal.RequireFromString("0.01")}
	tier := &models.TContractRiskLimitTier{MaintenanceMarginRate: decimal.RequireFromString("0.02"), MaintenanceAmount: decimal.NewFromInt(5)}
	recalculatePositionRisk(position, contract, tier)
	if !position.MaintenanceMargin.Equal(decimal.NewFromInt(15)) {
		t.Fatalf("tier maintenance = %s, want 15", position.MaintenanceMargin)
	}

	position.MarginMode = int64(trade.MarginMode_MARGIN_MODE_CROSS)
	recalculatePositionRisk(position, contract, tier)
	if !position.RiskRate.IsZero() || !position.LiquidationPrice.IsZero() || !position.BankruptcyPrice.IsZero() {
		t.Fatal("cross position must not use isolated liquidation approximation")
	}
}

func TestMarkRiskProjectionEqual(t *testing.T) {
	before := &models.TContractPosition{
		MarkPrice:         decimal.NewFromInt(110),
		MarkSnapshotId:    "mark-v1",
		MaintenanceMargin: decimal.RequireFromString("3.4"),
		UnrealizedPnl:     decimal.NewFromInt(20),
		LiquidationPrice:  decimal.RequireFromString("81.1224489795918367"),
		BankruptcyPrice:   decimal.NewFromInt(80),
		RiskRate:          decimal.RequireFromString("0.0566666667"),
		AdlRank:           7,
	}
	after := *before
	if !markRiskProjectionEqual(before, &after) {
		t.Fatal("identical mark projection must not cause a version-only database write")
	}
	after.MaintenanceMargin = decimal.RequireFromString("3.5")
	if markRiskProjectionEqual(before, &after) {
		t.Fatal("changed tier-derived risk must still be persisted for the same mark")
	}
}

func TestADLPriorityRank(t *testing.T) {
	profitable := &models.TContractPosition{
		MarginMode:     int64(trade.MarginMode_MARGIN_MODE_ISOLATED),
		PositionMargin: decimal.NewFromInt(20),
		UnrealizedPnl:  decimal.NewFromInt(20),
	}
	if got := adlPriorityRank(profitable, decimal.NewFromInt(120)); got != 3_000_000 {
		t.Fatalf("ADL rank = %d, want 3000000", got)
	}

	lessProfitable := *profitable
	lessProfitable.UnrealizedPnl = decimal.NewFromInt(10)
	if got := adlPriorityRank(&lessProfitable, decimal.NewFromInt(110)); got != 1_833_333 {
		t.Fatalf("lower-profit ADL rank = %d, want 1833333", got)
	}

	cross := *profitable
	cross.MarginMode = int64(trade.MarginMode_MARGIN_MODE_CROSS)
	if got := adlPriorityRank(&cross, decimal.NewFromInt(120)); got != 0 {
		t.Fatalf("cross position ADL rank = %d, want 0", got)
	}

	profitable.UnrealizedPnl = decimal.NewFromInt(-1)
	if got := adlPriorityRank(profitable, decimal.NewFromInt(120)); got != 0 {
		t.Fatalf("losing position ADL rank = %d, want 0", got)
	}
}
