package helpers

import (
	"context"
	"testing"

	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func TestUserTradeControlUsesExposureInsteadOfBuyEqualsOpen(t *testing.T) {
	limit := &models.TRiskUserTradeLimit{
		ControlMode: int64(trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_CLOSE_ONLY),
		CanOpen:     1, CanClose: 1, CanCancel: 1,
		TradeEnabled: int64(common.Enable_ENABLE_ENABLED),
		Enabled:      int64(common.Enable_ENABLE_ENABLED),
	}
	symbol := &models.TTradeSymbol{ProductType: int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE)}

	opening, err := evaluateUserOrderRisk(context.Background(), nil, &trade.CheckOrderRiskReq{
		Side: common.Side_SIDE_SELL, ExposureIncreasing: common.YesNo_YES_NO_YES,
	}, symbol, limit, nil)
	if err != nil {
		t.Fatal(err)
	}
	if opening.Passed || opening.RejectCode != "REDUCE_ONLY" {
		t.Fatalf("opening decision=%+v", opening)
	}

	closing, err := evaluateUserOrderRisk(context.Background(), nil, &trade.CheckOrderRiskReq{
		Side: common.Side_SIDE_BUY, ExposureIncreasing: common.YesNo_YES_NO_NO,
	}, symbol, limit, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !closing.Passed {
		t.Fatalf("BUY close should pass; decision=%+v", closing)
	}
}

func TestUserTradeControlSymbolOverrideAndChannelFlags(t *testing.T) {
	product := &models.TRiskUserTradeLimit{
		ControlMode: int64(trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_NORMAL),
		CanOpen:     1, CanClose: 1, CanCancel: 1, CanTriggerOrder: 0, CanApiTrade: 0,
		TradeEnabled: int64(common.Enable_ENABLE_ENABLED),
		Enabled:      int64(common.Enable_ENABLE_ENABLED),
	}
	symbol := &models.TTradeSymbol{ProductType: int64(common.ProductType_PRODUCT_TYPE_SPOT)}

	apiDecision, err := evaluateUserOrderRisk(context.Background(), nil, &trade.CheckOrderRiskReq{
		OrderSource: trade.OrderSourceType_ORDER_SOURCE_TYPE_API,
	}, symbol, product, nil)
	if err != nil {
		t.Fatal(err)
	}
	if apiDecision.Passed || apiDecision.RejectCode != "API_TRADE_DISABLED" {
		t.Fatalf("API decision=%+v", apiDecision)
	}

	override := &models.TRiskUserSymbolLimit{
		ControlMode: int64(trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_DISABLED),
		Enabled:     int64(common.Enable_ENABLE_ENABLED),
	}
	disabled, err := evaluateUserOrderRisk(context.Background(), nil, &trade.CheckOrderRiskReq{}, symbol, product, override)
	if err != nil {
		t.Fatal(err)
	}
	if disabled.Passed || disabled.RejectCode != "TRADE_DISABLED" {
		t.Fatalf("symbol override decision=%+v", disabled)
	}
}

func TestUserTradeControlReduceOnlyRequiresExplicitDerivativeFlag(t *testing.T) {
	limit := &models.TRiskUserTradeLimit{
		ControlMode: int64(trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_REDUCE_ONLY),
		CanOpen:     1, CanClose: 1, TradeEnabled: int64(common.Enable_ENABLE_ENABLED), Enabled: int64(common.Enable_ENABLE_ENABLED),
	}
	symbol := &models.TTradeSymbol{ProductType: int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE)}
	implicit, err := evaluateUserOrderRisk(context.Background(), nil, &trade.CheckOrderRiskReq{
		ExposureIncreasing: common.YesNo_YES_NO_NO, IsReduceOnly: common.YesNo_YES_NO_NO,
	}, symbol, limit, nil)
	if err != nil {
		t.Fatal(err)
	}
	if implicit.Passed || implicit.RejectCode != "REDUCE_ONLY" {
		t.Fatalf("implicit reduce decision=%+v", implicit)
	}
	explicit, err := evaluateUserOrderRisk(context.Background(), nil, &trade.CheckOrderRiskReq{
		ExposureIncreasing: common.YesNo_YES_NO_NO, IsReduceOnly: common.YesNo_YES_NO_YES,
	}, symbol, limit, nil)
	if err != nil || !explicit.Passed {
		t.Fatalf("explicit reduce decision=%+v err=%v", explicit, err)
	}
}

func TestUserTradeControlPriceDeviation(t *testing.T) {
	symbol := &models.TTradeSymbol{ProductType: int64(common.ProductType_PRODUCT_TYPE_SPOT)}
	limit := &models.TRiskUserSymbolLimit{
		ControlMode: int64(trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_NORMAL),
		Enabled:     int64(common.Enable_ENABLE_ENABLED), PriceDeviationRate: MustParseFloat("0.05"),
	}
	decision, err := evaluateUserOrderRisk(context.Background(), nil, &trade.CheckOrderRiskReq{
		OrderType: trade.OrderType_ORDER_TYPE_LIMIT, Price: "106", ReferencePrice: "100",
	}, symbol, nil, limit)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Passed || decision.RejectCode != "PRICE_DEVIATION" {
		t.Fatalf("price decision=%+v", decision)
	}
}

func TestUserTradeControlEffectiveWindow(t *testing.T) {
	item := &models.TRiskUserTradeLimit{Enabled: int64(common.Enable_ENABLE_ENABLED), EffectiveStartTime: 100, EffectiveEndTime: 200}
	if isProductLimitEffective(item, 99) || !isProductLimitEffective(item, 100) || isProductLimitEffective(item, 200) {
		t.Fatal("effective window must be [start,end)")
	}
}

func TestSpotPositionExposureLimits(t *testing.T) {
	quantity := evaluateSpotPositionExposure(
		MustParseFloat("9"), MustParseFloat("2"), MustParseFloat("200"), MustParseFloat("100"),
		MustParseFloat("100"), MustParseFloat("10"), decimal.Zero,
	)
	if quantity.Passed || quantity.RejectCode != "MAX_POSITION_QTY" {
		t.Fatalf("quantity decision=%+v", quantity)
	}

	notional := evaluateSpotPositionExposure(
		MustParseFloat("5"), MustParseFloat("1"), MustParseFloat("100"), MustParseFloat("100"),
		MustParseFloat("100"), decimal.Zero, MustParseFloat("550"),
	)
	if notional.Passed || notional.RejectCode != "MAX_POSITION_NOTIONAL" {
		t.Fatalf("notional decision=%+v", notional)
	}

	pass := evaluateSpotPositionExposure(
		MustParseFloat("3"), MustParseFloat("1"), MustParseFloat("100"), MustParseFloat("100"),
		MustParseFloat("100"), MustParseFloat("10"), MustParseFloat("1000"),
	)
	if !pass.Passed {
		t.Fatalf("pass decision=%+v", pass)
	}
}
