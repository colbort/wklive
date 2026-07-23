package tasklogic

import (
	"context"
	"database/sql"
	"testing"

	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

func testDecimal(value int64) decimal.Decimal { return decimal.NewFromInt(value) }

func TestOrderStatusAfterFill(t *testing.T) {
	tests := []struct {
		name  string
		order *models.TTradeOrder
		want  int64
	}{
		{
			name:  "no fill remains pending",
			order: &models.TTradeOrder{Qty: testDecimal(10), Amount: testDecimal(10000)},
			want:  int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		},
		{
			name:  "partial by qty",
			order: &models.TTradeOrder{Qty: testDecimal(10), Amount: testDecimal(10000), FilledQty: testDecimal(4), FilledAmount: testDecimal(4000)},
			want:  int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED),
		},
		{
			name:  "filled by qty",
			order: &models.TTradeOrder{Qty: testDecimal(10), Amount: testDecimal(10000), FilledQty: testDecimal(10), FilledAmount: testDecimal(10000)},
			want:  int64(trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING),
		},
		{
			name:  "filled by amount when qty target missing",
			order: &models.TTradeOrder{Amount: testDecimal(10000), FilledAmount: testDecimal(10000)},
			want:  int64(trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := orderStatusAfterFill(tt.order); got != tt.want {
				t.Fatalf("orderStatusAfterFill() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestOrderStateCategories(t *testing.T) {
	if !isOpenOrderStatus(int64(trade.OrderStatus_ORDER_STATUS_PENDING)) {
		t.Fatal("pending should be open")
	}
	if !isOpenOrderStatus(int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED)) {
		t.Fatal("part-filled should be open")
	}
	if !isOpenOrderStatus(int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING)) {
		t.Fatal("trigger-waiting should be open")
	}
	if isMatchableOrderStatus(int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING)) {
		t.Fatal("trigger-waiting should not be matchable")
	}
	if isOpenOrderStatus(int64(trade.OrderStatus_ORDER_STATUS_CANCELED)) {
		t.Fatal("canceled should not be open")
	}
	if isOpenOrderStatus(int64(trade.OrderStatus_ORDER_STATUS_FREEZING)) {
		t.Fatal("freezing should not be open")
	}
	if !isTerminalOrderStatus(int64(trade.OrderStatus_ORDER_STATUS_FILLED)) {
		t.Fatal("filled should be terminal")
	}
	if !isTerminalOrderStatus(int64(trade.OrderStatus_ORDER_STATUS_EXPIRED)) {
		t.Fatal("expired should be terminal")
	}
	for _, status := range []trade.OrderStatus{trade.OrderStatus_ORDER_STATUS_CANCELING, trade.OrderStatus_ORDER_STATUS_EXPIRING, trade.OrderStatus_ORDER_STATUS_SETTLEMENT_PENDING} {
		if isMatchableOrderStatus(int64(status)) || isTerminalOrderStatus(int64(status)) {
			t.Fatalf("intermediate status %s must be neither matchable nor terminal", status)
		}
	}
}

func TestShouldExpireOrder(t *testing.T) {
	now := int64(120_000)

	if !shouldExpireOrder(&models.TTradeOrder{
		Status:      int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_IOC),
		CreateTimes: now - immediateOrderExpireDelayMillis,
	}, now) {
		t.Fatal("old IOC order should expire")
	}

	if shouldExpireOrder(&models.TTradeOrder{
		Status:      int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_GTC),
		CreateTimes: now - immediateOrderExpireDelayMillis,
	}, now) {
		t.Fatal("GTC order should not expire")
	}

	if shouldExpireOrder(&models.TTradeOrder{
		Status:      int64(trade.OrderStatus_ORDER_STATUS_FILLED),
		TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_FOK),
		CreateTimes: now - immediateOrderExpireDelayMillis,
	}, now) {
		t.Fatal("terminal order should not expire")
	}

	triggerExt, err := marshalOrderAssetExt(orderAssetExt{TriggeredAt: now})
	if err != nil {
		t.Fatal(err)
	}
	if shouldExpireOrder(&models.TTradeOrder{
		Status:      int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_IOC),
		CreateTimes: now - immediateOrderExpireDelayMillis,
		BizExt:      sql.NullString{String: triggerExt, Valid: true},
	}, now) {
		t.Fatal("freshly triggered IOC order should not expire by original create time")
	}

	oldTriggerExt, err := marshalOrderAssetExt(orderAssetExt{TriggeredAt: now - immediateOrderExpireDelayMillis})
	if err != nil {
		t.Fatal(err)
	}
	if !shouldExpireOrder(&models.TTradeOrder{
		Status:      int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_IOC),
		CreateTimes: now - immediateOrderExpireDelayMillis*2,
		BizExt:      sql.NullString{String: oldTriggerExt, Valid: true},
	}, now) {
		t.Fatal("old triggered IOC order should expire by triggered time")
	}
}

func TestShouldRecoverFreezingOrder(t *testing.T) {
	now := int64(120_000)
	if !shouldRecoverFreezingOrder(&models.TTradeOrder{
		Status:      int64(trade.OrderStatus_ORDER_STATUS_FREEZING),
		CreateTimes: now - freezingOrderRecoverDelayMillis,
	}, now) {
		t.Fatal("old freezing order should recover")
	}
	if shouldRecoverFreezingOrder(&models.TTradeOrder{
		Status:      int64(trade.OrderStatus_ORDER_STATUS_FREEZING),
		CreateTimes: now - freezingOrderRecoverDelayMillis + 1,
	}, now) {
		t.Fatal("new freezing order should not recover")
	}
	if shouldRecoverFreezingOrder(&models.TTradeOrder{
		Status:      int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes: now - freezingOrderRecoverDelayMillis,
	}, now) {
		t.Fatal("non-freezing order should not recover")
	}
}

func TestOrderInputGuards(t *testing.T) {
	if isValidOrderPrice(trade.OrderType_ORDER_TYPE_LIMIT, decimal.Zero) {
		t.Fatal("limit order without price should be invalid")
	}
	if !hasNegativeOrderInput(decimal.Zero, testDecimal(1), testDecimal(-1), decimal.Zero) {
		t.Fatal("negative order amount should be invalid")
	}
	if !isValidOrderPrice(trade.OrderType_ORDER_TYPE_MARKET, decimal.Zero) {
		t.Fatal("market order should not require user price")
	}
	if isValidOrderTimeInForce(trade.OrderType_ORDER_TYPE_MARKET, trade.TriggerKind_TRIGGER_KIND_NONE, trade.TimeInForce_TIME_IN_FORCE_POST_ONLY) {
		t.Fatal("market post-only should be invalid")
	}
	if got := normalizeOrderTimeInForce(trade.OrderType_ORDER_TYPE_LIMIT, trade.TimeInForce_TIME_IN_FORCE_UNKNOWN); got != trade.TimeInForce_TIME_IN_FORCE_GTC {
		t.Fatalf("limit default TIF = %v, want GTC", got)
	}
	if got := normalizeOrderTimeInForce(trade.OrderType_ORDER_TYPE_MARKET, trade.TimeInForce_TIME_IN_FORCE_GTC); got != trade.TimeInForce_TIME_IN_FORCE_IOC {
		t.Fatalf("market GTC should normalize to IOC, got %v", got)
	}
	if isValidOrderTimeInForce(trade.OrderType_ORDER_TYPE_LIMIT, trade.TriggerKind_TRIGGER_KIND_STOP_LOSS, trade.TimeInForce_TIME_IN_FORCE_POST_ONLY) {
		t.Fatal("trigger order post-only should be invalid")
	}
}

func TestMatchExecutionPrice(t *testing.T) {
	tests := []struct {
		name string
		buy  *models.TTradeOrder
		sell *models.TTradeOrder
		want decimal.Decimal
		ok   bool
	}{
		{
			name: "limit orders crossed use maker price",
			buy:  &models.TTradeOrder{Id: 1, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(101)},
			sell: &models.TTradeOrder{Id: 2, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(100)},
			want: testDecimal(101),
			ok:   true,
		},
		{
			name: "limit orders not crossed",
			buy:  &models.TTradeOrder{Id: 1, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(99)},
			sell: &models.TTradeOrder{Id: 2, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(100)},
			ok:   false,
		},
		{
			name: "market buy uses sell price",
			buy:  &models.TTradeOrder{Id: 2, OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET)},
			sell: &models.TTradeOrder{Id: 1, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(100)},
			want: testDecimal(100),
			ok:   true,
		},
		{
			name: "two market orders cannot price",
			buy:  &models.TTradeOrder{Id: 1, OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET)},
			sell: &models.TTradeOrder{Id: 2, OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET)},
			ok:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := matchExecutionPrice(tt.buy, tt.sell)
			if ok != tt.ok {
				t.Fatalf("matchExecutionPrice() ok = %v, want %v", ok, tt.ok)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("matchExecutionPrice() price = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectOrderMatchPlanSkipsMarketMarketPair(t *testing.T) {
	buys := []*models.TTradeOrder{
		{Id: 1, Side: int64(common.Side_SIDE_BUY), OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET), Qty: testDecimal(1)},
	}
	sells := []*models.TTradeOrder{
		{Id: 2, Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET), Qty: testDecimal(1)},
		{Id: 3, Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(10), Qty: testDecimal(1)},
	}
	plan := selectOrderMatchPlan(buys, sells)
	if plan == nil {
		t.Fatal("market buy should match the sell limit behind a sell market order")
	}
	if plan.BuyOrder.Id != 1 || plan.SellOrder.Id != 3 || !plan.Price.Equal(testDecimal(10)) {
		t.Fatalf("selected plan = buy %d sell %d price %v, want buy 1 sell 3 price 10", plan.BuyOrder.Id, plan.SellOrder.Id, plan.Price)
	}
}

func TestSelectOrderMatchPlanKeepsBookPriority(t *testing.T) {
	buys := []*models.TTradeOrder{
		{Id: 100, Side: int64(common.Side_SIDE_BUY), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(100), Qty: testDecimal(1)},
		{Id: 1, Side: int64(common.Side_SIDE_BUY), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(99), Qty: testDecimal(1)},
	}
	sells := []*models.TTradeOrder{
		{Id: 2, Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(90), Qty: testDecimal(1)},
	}
	plan := selectOrderMatchPlan(buys, sells)
	if plan == nil {
		t.Fatal("crossed book should select a match")
	}
	if plan.BuyOrder.Id != 100 {
		t.Fatalf("selected buy order = %d, want highest-priority buy 100", plan.BuyOrder.Id)
	}
}

func TestRemainingMatchQty(t *testing.T) {
	got := remainingMatchQty(&models.TTradeOrder{Amount: testDecimal(100), FilledAmount: testDecimal(25)}, testDecimal(5))
	if !got.Equal(testDecimal(15)) {
		t.Fatalf("remainingMatchQty() = %v, want 15", got)
	}

	got = remainingMatchQty(&models.TTradeOrder{Qty: testDecimal(10), FilledQty: testDecimal(4)}, testDecimal(5))
	if !got.Equal(testDecimal(6)) {
		t.Fatalf("remainingMatchQty() = %v, want 6", got)
	}
}

func TestOrderFillNeedByAmountUsesNaturalAmount(t *testing.T) {
	need := orderFillNeed{remainingAmount: testDecimal(100)}
	if got := need.matchQty(testDecimal(20)); !got.Equal(testDecimal(5)) {
		t.Fatalf("matchQty() = %v, want 5", got)
	}
}

func TestCanApplyOrderFill(t *testing.T) {
	qtyOrder := &models.TTradeOrder{
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Side:      int64(common.Side_SIDE_SELL),
		Price:     testDecimal(10),
		Qty:       testDecimal(10),
		FilledQty: testDecimal(4),
		Amount:    testDecimal(100),
	}
	if !canApplyOrderFill(qtyOrder, &models.TTradeFill{Price: testDecimal(12), Qty: testDecimal(6), Amount: testDecimal(72)}) {
		t.Fatal("remaining qty should be fillable")
	}
	if canApplyOrderFill(qtyOrder, &models.TTradeFill{Price: testDecimal(12), Qty: testDecimal(7), Amount: testDecimal(84)}) {
		t.Fatal("fill should not exceed remaining qty")
	}
	if canApplyOrderFill(qtyOrder, &models.TTradeFill{Price: testDecimal(9), Qty: testDecimal(6), Amount: testDecimal(54)}) {
		t.Fatal("sell limit fill price should not be below order price")
	}

	amountOrder := &models.TTradeOrder{Amount: testDecimal(100), FilledAmount: testDecimal(40)}
	if !canApplyOrderFill(amountOrder, &models.TTradeFill{Qty: testDecimal(3), Amount: testDecimal(60)}) {
		t.Fatal("remaining amount should be fillable")
	}
	if canApplyOrderFill(amountOrder, &models.TTradeFill{Qty: testDecimal(3), Amount: testDecimal(61)}) {
		t.Fatal("fill should not exceed remaining amount")
	}
}

func TestFillMatchesOrder(t *testing.T) {
	order := &models.TTradeOrder{OrderNo: "TRD1", UserId: 1, SymbolId: 2, ProductType: 1, Side: int64(common.Side_SIDE_BUY)}
	if !fillMatchesOrder(order, &models.TTradeFill{OrderNo: "TRD1", UserId: 1, SymbolId: 2, ProductType: 1, Side: int64(common.Side_SIDE_BUY)}) {
		t.Fatal("matching fill metadata should pass")
	}
	if fillMatchesOrder(order, &models.TTradeFill{OrderNo: "TRD2", UserId: 1, SymbolId: 2, ProductType: 1, Side: int64(common.Side_SIDE_BUY)}) {
		t.Fatal("mismatched order no should fail")
	}
}

func TestShouldTriggerOrder(t *testing.T) {
	base := models.TTradeOrder{
		Status:       int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING),
		TriggerPrice: testDecimal(100),
	}
	tests := []struct {
		name  string
		order models.TTradeOrder
		price decimal.Decimal
		want  bool
	}{
		{
			name:  "sell take profit triggers upward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT), Side: int64(common.Side_SIDE_SELL)},
			price: testDecimal(101),
			want:  true,
		},
		{
			name:  "sell stop loss triggers downward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_STOP_LOSS), Side: int64(common.Side_SIDE_SELL)},
			price: testDecimal(99),
			want:  true,
		},
		{
			name:  "buy take profit triggers downward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT), Side: int64(common.Side_SIDE_BUY)},
			price: testDecimal(99),
			want:  true,
		},
		{
			name:  "buy stop loss triggers upward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_STOP_LOSS), Side: int64(common.Side_SIDE_BUY)},
			price: testDecimal(101),
			want:  true,
		},
		{
			name:  "non waiting order does not trigger",
			order: models.TTradeOrder{Status: int64(trade.OrderStatus_ORDER_STATUS_PENDING), TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_STOP_LOSS), Side: int64(common.Side_SIDE_SELL)},
			price: testDecimal(99),
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldTriggerOrder(&tt.order, tt.price); got != tt.want {
				t.Fatalf("shouldTriggerOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTriggeredOrderExecutionType(t *testing.T) {
	if got := triggeredOrderExecutionType(&models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(10)}); got != int64(trade.OrderType_ORDER_TYPE_LIMIT) {
		t.Fatalf("triggeredOrderExecutionType() = %d, want LIMIT", got)
	}
	if got := triggeredOrderExecutionType(&models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET)}); got != int64(trade.OrderType_ORDER_TYPE_MARKET) {
		t.Fatalf("triggeredOrderExecutionType() = %d, want MARKET", got)
	}
	if got := triggeredOrderExecutionType(&models.TTradeOrder{OrderType: legacyOrderTypeStopLoss, Price: testDecimal(10)}); got != int64(trade.OrderType_ORDER_TYPE_LIMIT) {
		t.Fatalf("legacy triggered order execution type = %d, want LIMIT", got)
	}
	if got := triggeredTimeInForce(&models.TTradeOrder{}); got != int64(trade.TimeInForce_TIME_IN_FORCE_IOC) {
		t.Fatalf("triggeredTimeInForce() = %d, want IOC", got)
	}
}

func TestPostOnlyWouldTake(t *testing.T) {
	if !postOnlyWouldTake(
		&models.TTradeOrder{Id: 2, TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_POST_ONLY)},
		&models.TTradeOrder{Id: 1},
	) {
		t.Fatal("newer post-only buy should be taker")
	}
	if postOnlyWouldTake(
		&models.TTradeOrder{Id: 1, TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_POST_ONLY)},
		&models.TTradeOrder{Id: 2},
	) {
		t.Fatal("older post-only buy should be maker")
	}
}

func TestCanFullyFillFromBook(t *testing.T) {
	order := &models.TTradeOrder{
		Side:      int64(common.Side_SIDE_BUY),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Price:     testDecimal(100),
		Qty:       testDecimal(10),
	}
	opposites := []*models.TTradeOrder{
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(99), Qty: testDecimal(4)},
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(100), Qty: testDecimal(6)},
	}
	if !canFullyFillFromBook(order, opposites) {
		t.Fatal("order should be fully fillable across multiple levels")
	}

	largeFOK := []*models.TTradeOrder{
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_FOK), Price: testDecimal(99), Qty: testDecimal(20)},
	}
	if canFullyFillFromBook(order, largeFOK) {
		t.Fatal("order should not partially fill an opposite FOK order")
	}

	withSkippedFOK := append(largeFOK, &models.TTradeOrder{
		Side:      int64(common.Side_SIDE_SELL),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Price:     testDecimal(100),
		Qty:       testDecimal(10),
	})
	if !canFullyFillFromBook(order, withSkippedFOK) {
		t.Fatal("order should skip incompatible FOK liquidity and fill from the next level")
	}
}

func TestCanFullyFillFromBookRespectsPostOnly(t *testing.T) {
	order := &models.TTradeOrder{
		Id:          1,
		Side:        int64(common.Side_SIDE_BUY),
		OrderType:   int64(trade.OrderType_ORDER_TYPE_LIMIT),
		TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_FOK),
		Price:       testDecimal(100),
		Qty:         testDecimal(10),
	}
	opposites := []*models.TTradeOrder{
		{Id: 2, Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_POST_ONLY), Price: testDecimal(99), Qty: testDecimal(10)},
	}
	if canFullyFillFromBook(order, opposites) {
		t.Fatal("post-only liquidity that would take should not satisfy FOK")
	}
}

func TestCanFullyFillFromBookByAmount(t *testing.T) {
	order := &models.TTradeOrder{
		Side:      int64(common.Side_SIDE_BUY),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Price:     testDecimal(100),
		Amount:    testDecimal(100),
	}
	opposites := []*models.TTradeOrder{
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(10), Qty: testDecimal(5)},
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: testDecimal(20), Qty: decimal.RequireFromString("2.5")},
	}
	if !canFullyFillFromBook(order, opposites) {
		t.Fatal("amount based order should be fillable by accumulated turnover")
	}
}

func TestResidualExpireReason(t *testing.T) {
	if got := residualExpireReason(&models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET)}, nil, nil); got == "" {
		t.Fatal("market residual should expire")
	}
	if got := residualExpireReason(&models.TTradeOrder{TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_IOC)}, nil, nil); got == "" {
		t.Fatal("IOC residual should expire")
	}
	if got := residualExpireReason(&models.TTradeOrder{TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_FOK)}, nil, nil); got == "" {
		t.Fatal("FOK residual should expire")
	}
}

func TestTradeFillFromProtoRequiresCompleteExecution(t *testing.T) {
	if _, err := tradeFillFromProto(&trade.TradeFill{
		TenantId: 1,
		FillNo:   "FIL1",
		OrderId:  1,
		Qty:      "1",
	}, 1); err == nil {
		t.Fatal("fill without positive price and amount should be rejected")
	}

	fill, err := tradeFillFromProto(&trade.TradeFill{
		TenantId: 1,
		FillNo:   "FIL2",
		MatchNo:  "MAT1",
		OrderId:  1,
		Price:    "10",
		Qty:      "2",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !fill.Amount.Equal(testDecimal(20)) {
		t.Fatalf("computed fill amount = %v, want 20", fill.Amount)
	}
}

func TestBuildSpotFillSettlementInstructions(t *testing.T) {
	symbol := &models.TTradeSymbol{BaseAsset: "BTC", QuoteAsset: "USDT"}
	fill := &models.TTradeFill{Qty: testDecimal(2), Amount: testDecimal(20000), Fee: testDecimal(10), FeeAsset: "USDT"}

	buy, err := buildFillSettlementInstructions(context.Background(), nil, symbol, &models.TTradeOrder{ProductType: int64(trade.ProductType_PRODUCT_TYPE_SPOT), Side: int64(common.Side_SIDE_BUY)}, fill)
	if err != nil {
		t.Fatal(err)
	}
	if len(buy) != 3 || buy[0].asset != "USDT" || buy[0].action != trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CONSUME_FROZEN || !buy[0].amount.Equal(fill.Amount) || buy[1].asset != "BTC" || !buy[1].amount.Equal(toTradeMinorAmount(fill.Qty)) {
		t.Fatalf("unexpected spot buy settlement instructions: %+v", buy)
	}

	sell, err := buildFillSettlementInstructions(context.Background(), nil, symbol, &models.TTradeOrder{ProductType: int64(trade.ProductType_PRODUCT_TYPE_SPOT), Side: int64(common.Side_SIDE_SELL)}, fill)
	if err != nil {
		t.Fatal(err)
	}
	if len(sell) != 3 || sell[0].asset != "BTC" || !sell[0].amount.Equal(toTradeMinorAmount(fill.Qty)) || sell[1].asset != "USDT" || !sell[1].amount.Equal(fill.Amount) {
		t.Fatalf("unexpected spot sell settlement instructions: %+v", sell)
	}
}

func TestOrderBookKeyAndMember(t *testing.T) {
	order := &models.TTradeOrder{
		Id:          123,
		TenantId:    7,
		SymbolId:    4,
		ProductType: int64(trade.ProductType_PRODUCT_TYPE_DERIVATIVE),
		Side:        int64(common.Side_SIDE_BUY),
	}
	if got, want := orderBookKey(order), "trade:book:7:2:4:buy"; got != want {
		t.Fatalf("orderBookKey() = %q, want %q", got, want)
	}
	member := orderBookMember(order.Id)
	if got, err := orderBookMemberID(member); err != nil || got != order.Id {
		t.Fatalf("orderBookMemberID() = %d, err = %v, want %d", got, err, order.Id)
	}
}

func TestOrderBookScorePriority(t *testing.T) {
	marketBuy := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET), Side: int64(common.Side_SIDE_BUY)}
	highBuy := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Side: int64(common.Side_SIDE_BUY), Price: testDecimal(101)}
	lowBuy := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Side: int64(common.Side_SIDE_BUY), Price: testDecimal(100)}
	lowSell := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Side: int64(common.Side_SIDE_SELL), Price: testDecimal(100)}
	highSell := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Side: int64(common.Side_SIDE_SELL), Price: testDecimal(101)}

	if !(orderBookScore(marketBuy) < orderBookScore(highBuy)) {
		t.Fatal("market order should rank before limit orders")
	}
	if !(orderBookScore(highBuy) < orderBookScore(lowBuy)) {
		t.Fatal("higher buy price should rank first in ascending zset order")
	}
	if !(orderBookScore(lowSell) < orderBookScore(highSell)) {
		t.Fatal("lower sell price should rank first in ascending zset order")
	}
}

func TestIsOrderBookOrder(t *testing.T) {
	if !isOrderBookOrder(&models.TTradeOrder{
		Status:    int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
	}) {
		t.Fatal("pending limit order should enter order book")
	}
	if isOrderBookOrder(&models.TTradeOrder{
		Status:    int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
	}) {
		t.Fatal("trigger waiting order should not enter order book")
	}
	if isOrderBookOrder(&models.TTradeOrder{
		Status:    int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		OrderType: legacyOrderTypeStopLoss,
	}) {
		t.Fatal("untriggered stop order type should not enter order book")
	}
}
