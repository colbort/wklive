package applogic

import (
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestConsumeBuyOrderReservation(t *testing.T) {
	order := &models.TOptionOrder{
		Side:           int64(common.Side_SIDE_BUY),
		PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		MarginAmount:   decimal.NewFromInt(100),
	}
	consumeBuyOrderReservation(order, decimal.NewFromInt(35))
	if !order.MarginAmount.Equal(decimal.NewFromInt(65)) {
		t.Fatalf("unexpected remaining reservation: %s", order.MarginAmount)
	}
	consumeBuyOrderReservation(order, decimal.NewFromInt(80))
	if !order.MarginAmount.IsZero() {
		t.Fatalf("reservation must never become negative: %s", order.MarginAmount)
	}

	sellOrder := &models.TOptionOrder{
		Side:           int64(common.Side_SIDE_SELL),
		PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_CLOSE),
		MarginAmount:   decimal.NewFromInt(100),
	}
	consumeBuyOrderReservation(sellOrder, decimal.NewFromInt(35))
	if !sellOrder.MarginAmount.Equal(decimal.NewFromInt(100)) {
		t.Fatal("sell order does not own a premium reservation")
	}

	buyCloseOrder := &models.TOptionOrder{
		Side:           int64(common.Side_SIDE_BUY),
		PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_CLOSE),
		MarginAmount:   decimal.NewFromInt(100),
	}
	consumeBuyOrderReservation(buyCloseOrder, decimal.NewFromInt(35))
	if !buyCloseOrder.MarginAmount.Equal(decimal.NewFromInt(65)) {
		t.Fatal("buy-to-close must consume its premium reservation")
	}
}

func TestRejectComboChildFromSimpleMatcher(t *testing.T) {
	if err := rejectComboChildFromSimpleMatcher(nil, "test"); err != nil {
		t.Fatalf("nil order rejected: %v", err)
	}
	if err := rejectComboChildFromSimpleMatcher(&models.TOptionOrder{
		TenantId: 9, ComboOrderId: 0,
	}, "test"); err != nil {
		t.Fatalf("simple order rejected: %v", err)
	}
	if err := rejectComboChildFromSimpleMatcher(&models.TOptionOrder{
		TenantId: 9, ComboOrderId: 99,
	}, "test"); err == nil {
		t.Fatal("combo shadow order must be rejected from the simple matcher")
	}
}

func TestMatchableOptionQtyExcludesSelfTrading(t *testing.T) {
	incoming := &models.TOptionOrder{Id: 10, UserId: 20, AccountId: 30}
	candidates := []*models.TOptionOrder{
		{Id: 10, UserId: 99, AccountId: 99, UnfilledQty: decimal.NewFromInt(7)},
		{Id: 11, UserId: 20, AccountId: 30, UnfilledQty: decimal.NewFromInt(5)},
		{Id: 12, UserId: 20, AccountId: 31, UnfilledQty: decimal.NewFromInt(3)},
		{Id: 13, UserId: 21, AccountId: 30, UnfilledQty: decimal.NewFromInt(2)},
		{Id: 14, UserId: 21, AccountId: 31, UnfilledQty: decimal.Zero},
	}
	if got := matchableOptionQty(incoming, candidates); !got.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("unexpected matchable quantity: %s", got)
	}
}

func TestImmediateOptionOrderTypes(t *testing.T) {
	for _, orderType := range []option.OrderType{
		option.OrderType_ORDER_TYPE_MARKET,
		option.OrderType_ORDER_TYPE_IOC,
		option.OrderType_ORDER_TYPE_FOK,
	} {
		if !isImmediateOptionOrder(int64(orderType)) {
			t.Fatalf("%s must cancel its unfilled remainder", orderType)
		}
	}
	for _, orderType := range []option.OrderType{
		option.OrderType_ORDER_TYPE_LIMIT,
		option.OrderType_ORDER_TYPE_POST_ONLY,
	} {
		if isImmediateOptionOrder(int64(orderType)) {
			t.Fatalf("%s must be allowed to rest on book", orderType)
		}
	}
}

func TestOppositeOrderSide(t *testing.T) {
	if got := oppositeOrderSide(int64(common.Side_SIDE_BUY)); got != int64(common.Side_SIDE_SELL) {
		t.Fatalf("buy opposite side = %d", got)
	}
	if got := oppositeOrderSide(int64(common.Side_SIDE_SELL)); got != int64(common.Side_SIDE_BUY) {
		t.Fatalf("sell opposite side = %d", got)
	}
}

func TestOptionTradeFeesUseMakerAndTakerRates(t *testing.T) {
	contract := &models.TOptionContract{
		MakerFeeRate: decimal.RequireFromString("0.001"),
		TakerFeeRate: decimal.RequireFromString("0.002"),
	}
	turnover := decimal.NewFromInt(1000)

	buyFee, sellFee := optionTradeFees(contract, turnover, int64(common.Side_SIDE_BUY))
	if !buyFee.Equal(decimal.NewFromInt(1)) || !sellFee.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("buy maker fees got buy=%s sell=%s", buyFee, sellFee)
	}

	buyFee, sellFee = optionTradeFees(contract, turnover, int64(common.Side_SIDE_SELL))
	if !buyFee.Equal(decimal.NewFromInt(2)) || !sellFee.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("sell maker fees got buy=%s sell=%s", buyFee, sellFee)
	}
}

func TestValidateOptionTradeAssetBalance(t *testing.T) {
	trade := &models.TOptionTrade{
		Turnover: decimal.NewFromInt(100),
		BuyFee:   decimal.NewFromInt(2),
		SellFee:  decimal.NewFromInt(1),
	}
	if err := validateOptionTradeAssetBalance(trade); err != nil {
		t.Fatalf("balanced trade rejected: %v", err)
	}
	trade.SellFee = decimal.NewFromInt(101)
	if err := validateOptionTradeAssetBalance(trade); err == nil {
		t.Fatal("seller fee above premium must be rejected")
	}
}

func TestOptionSellerMargin(t *testing.T) {
	contract := &models.TOptionContract{
		OptionType:            int64(option.OptionType_OPTION_TYPE_CALL),
		StrikePrice:           decimal.NewFromInt(120),
		Multiplier:            decimal.NewFromInt(1),
		InitialMarginRate:     decimal.RequireFromString("0.20"),
		MaintenanceMarginRate: decimal.RequireFromString("0.15"),
		MinMarginRate:         decimal.RequireFromString("0.10"),
	}
	initial := optionSellerMargin(contract, decimal.NewFromInt(100), decimal.NewFromInt(5), decimal.NewFromInt(2), false)
	maintenance := optionSellerMargin(contract, decimal.NewFromInt(100), decimal.NewFromInt(5), decimal.NewFromInt(2), true)
	if !initial.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("unexpected call initial margin: %s", initial)
	}
	if !maintenance.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("minimum margin floor should apply: %s", maintenance)
	}

	contract.OptionType = int64(option.OptionType_OPTION_TYPE_PUT)
	contract.StrikePrice = decimal.NewFromInt(80)
	put := optionSellerMargin(contract, decimal.NewFromInt(100), decimal.NewFromInt(4), decimal.NewFromInt(1), false)
	if !put.Equal(decimal.NewFromInt(8)) {
		t.Fatalf("unexpected OTM put margin: %s", put)
	}
}

func TestAllocateSellerMargin(t *testing.T) {
	order := &models.TOptionOrder{
		Side: int64(common.Side_SIDE_SELL), PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		MarginAmount: decimal.NewFromInt(100),
	}
	first := allocateSellerMargin(order, decimal.NewFromInt(3), decimal.NewFromInt(10))
	if !first.Equal(decimal.NewFromInt(30)) || !order.MarginAmount.Equal(decimal.NewFromInt(70)) {
		t.Fatalf("unexpected first margin allocation: allocated=%s remaining=%s", first, order.MarginAmount)
	}
	last := allocateSellerMargin(order, decimal.NewFromInt(7), decimal.NewFromInt(7))
	if !last.Equal(decimal.NewFromInt(70)) || !order.MarginAmount.IsZero() {
		t.Fatalf("final fill must allocate exact remainder: allocated=%s remaining=%s", last, order.MarginAmount)
	}
}
