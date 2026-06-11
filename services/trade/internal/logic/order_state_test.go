package logic

import (
	"database/sql"
	"testing"

	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/models"
)

func TestOrderStatusAfterFill(t *testing.T) {
	tests := []struct {
		name  string
		order *models.TTradeOrder
		want  int64
	}{
		{
			name:  "no fill remains pending",
			order: &models.TTradeOrder{Qty: 10, Amount: 10000},
			want:  int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		},
		{
			name:  "partial by qty",
			order: &models.TTradeOrder{Qty: 10, Amount: 10000, FilledQty: 4, FilledAmount: 4000},
			want:  int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED),
		},
		{
			name:  "filled by qty",
			order: &models.TTradeOrder{Qty: 10, Amount: 10000, FilledQty: 10, FilledAmount: 10000},
			want:  int64(trade.OrderStatus_ORDER_STATUS_FILLED),
		},
		{
			name:  "filled by amount when qty target missing",
			order: &models.TTradeOrder{Amount: 10000, FilledAmount: 10000},
			want:  int64(trade.OrderStatus_ORDER_STATUS_FILLED),
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
	if isValidOrderPrice(trade.OrderType_ORDER_TYPE_LIMIT, 0) {
		t.Fatal("limit order without price should be invalid")
	}
	if !hasNegativeOrderInput(0, 1, -1, 0) {
		t.Fatal("negative order amount should be invalid")
	}
	if !isValidOrderPrice(trade.OrderType_ORDER_TYPE_MARKET, 0) {
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

func TestOrderAmountPriceDistinguishesOrderType(t *testing.T) {
	logic := &PlaceOrderLogic{}
	if got, err := logic.orderAmountPrice(nil, trade.OrderType_ORDER_TYPE_LIMIT, 10); err != nil || got != 10 {
		t.Fatalf("limit amount price = %v, err = %v, want 10", got, err)
	}
	if got, err := logic.orderAmountPrice(nil, trade.OrderType_ORDER_TYPE_MARKET, 10); err != nil || got != 0 {
		t.Fatalf("market amount price = %v, err = %v, want market price lookup fallback 0", got, err)
	}
}

func TestMatchExecutionPrice(t *testing.T) {
	tests := []struct {
		name string
		buy  *models.TTradeOrder
		sell *models.TTradeOrder
		want float64
		ok   bool
	}{
		{
			name: "limit orders crossed use maker price",
			buy:  &models.TTradeOrder{Id: 1, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 101},
			sell: &models.TTradeOrder{Id: 2, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 100},
			want: 101,
			ok:   true,
		},
		{
			name: "limit orders not crossed",
			buy:  &models.TTradeOrder{Id: 1, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 99},
			sell: &models.TTradeOrder{Id: 2, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 100},
			ok:   false,
		},
		{
			name: "market buy uses sell price",
			buy:  &models.TTradeOrder{Id: 2, OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET)},
			sell: &models.TTradeOrder{Id: 1, OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 100},
			want: 100,
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
			if got != tt.want {
				t.Fatalf("matchExecutionPrice() price = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelectOrderMatchPlanSkipsMarketMarketPair(t *testing.T) {
	buys := []*models.TTradeOrder{
		{Id: 1, Side: int64(common.Side_SIDE_BUY), OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET), Qty: 1},
	}
	sells := []*models.TTradeOrder{
		{Id: 2, Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET), Qty: 1},
		{Id: 3, Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 10, Qty: 1},
	}
	plan := selectOrderMatchPlan(buys, sells)
	if plan == nil {
		t.Fatal("market buy should match the sell limit behind a sell market order")
	}
	if plan.BuyOrder.Id != 1 || plan.SellOrder.Id != 3 || plan.Price != 10 {
		t.Fatalf("selected plan = buy %d sell %d price %v, want buy 1 sell 3 price 10", plan.BuyOrder.Id, plan.SellOrder.Id, plan.Price)
	}
}

func TestSelectOrderMatchPlanKeepsBookPriority(t *testing.T) {
	buys := []*models.TTradeOrder{
		{Id: 100, Side: int64(common.Side_SIDE_BUY), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 100, Qty: 1},
		{Id: 1, Side: int64(common.Side_SIDE_BUY), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 99, Qty: 1},
	}
	sells := []*models.TTradeOrder{
		{Id: 2, Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 90, Qty: 1},
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
	got := remainingMatchQty(&models.TTradeOrder{Amount: 10000, FilledAmount: 2500}, 5)
	if got != 15 {
		t.Fatalf("remainingMatchQty() = %v, want 15", got)
	}

	got = remainingMatchQty(&models.TTradeOrder{Qty: 10, FilledQty: 4}, 5)
	if got != 6 {
		t.Fatalf("remainingMatchQty() = %v, want 6", got)
	}
}

func TestOrderFillNeedByAmountUsesMinorAmount(t *testing.T) {
	need := orderFillNeed{remainingAmount: 10000}
	if got := need.matchQty(20); got != 5 {
		t.Fatalf("matchQty() = %v, want 5", got)
	}
}

func TestCanApplyOrderFill(t *testing.T) {
	qtyOrder := &models.TTradeOrder{
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Side:      int64(common.Side_SIDE_SELL),
		Price:     10,
		Qty:       10,
		FilledQty: 4,
		Amount:    10000,
	}
	if !canApplyOrderFill(qtyOrder, &models.TTradeFill{Price: 12, Qty: 6, Amount: 7200}) {
		t.Fatal("remaining qty should be fillable")
	}
	if canApplyOrderFill(qtyOrder, &models.TTradeFill{Price: 12, Qty: 7, Amount: 8400}) {
		t.Fatal("fill should not exceed remaining qty")
	}
	if canApplyOrderFill(qtyOrder, &models.TTradeFill{Price: 9, Qty: 6, Amount: 5400}) {
		t.Fatal("sell limit fill price should not be below order price")
	}

	amountOrder := &models.TTradeOrder{Amount: 10000, FilledAmount: 4000}
	if !canApplyOrderFill(amountOrder, &models.TTradeFill{Qty: 3, Amount: 6000}) {
		t.Fatal("remaining amount should be fillable")
	}
	if canApplyOrderFill(amountOrder, &models.TTradeFill{Qty: 3, Amount: 6100}) {
		t.Fatal("fill should not exceed remaining amount")
	}
}

func TestFillMatchesOrder(t *testing.T) {
	order := &models.TTradeOrder{OrderNo: "TRD1", UserId: 1, SymbolId: 2, MarketType: 1, Side: int64(common.Side_SIDE_BUY)}
	if !fillMatchesOrder(order, &models.TTradeFill{OrderNo: "TRD1", UserId: 1, SymbolId: 2, MarketType: 1, Side: int64(common.Side_SIDE_BUY)}) {
		t.Fatal("matching fill metadata should pass")
	}
	if fillMatchesOrder(order, &models.TTradeFill{OrderNo: "TRD2", UserId: 1, SymbolId: 2, MarketType: 1, Side: int64(common.Side_SIDE_BUY)}) {
		t.Fatal("mismatched order no should fail")
	}
}

func TestShouldTriggerOrder(t *testing.T) {
	base := models.TTradeOrder{
		Status:       int64(trade.OrderStatus_ORDER_STATUS_TRIGGER_WAITING),
		TriggerPrice: 100,
	}
	tests := []struct {
		name  string
		order models.TTradeOrder
		price float64
		want  bool
	}{
		{
			name:  "sell take profit triggers upward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT), Side: int64(common.Side_SIDE_SELL)},
			price: 101,
			want:  true,
		},
		{
			name:  "sell stop loss triggers downward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_STOP_LOSS), Side: int64(common.Side_SIDE_SELL)},
			price: 99,
			want:  true,
		},
		{
			name:  "buy take profit triggers downward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_TAKE_PROFIT), Side: int64(common.Side_SIDE_BUY)},
			price: 99,
			want:  true,
		},
		{
			name:  "buy stop loss triggers upward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_STOP_LOSS), Side: int64(common.Side_SIDE_BUY)},
			price: 101,
			want:  true,
		},
		{
			name:  "non waiting order does not trigger",
			order: models.TTradeOrder{Status: int64(trade.OrderStatus_ORDER_STATUS_PENDING), TriggerPrice: base.TriggerPrice, TriggerKind: int64(trade.TriggerKind_TRIGGER_KIND_STOP_LOSS), Side: int64(common.Side_SIDE_SELL)},
			price: 99,
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
	if got := triggeredOrderExecutionType(&models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 10}); got != int64(trade.OrderType_ORDER_TYPE_LIMIT) {
		t.Fatalf("triggeredOrderExecutionType() = %d, want LIMIT", got)
	}
	if got := triggeredOrderExecutionType(&models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET)}); got != int64(trade.OrderType_ORDER_TYPE_MARKET) {
		t.Fatalf("triggeredOrderExecutionType() = %d, want MARKET", got)
	}
	if got := triggeredOrderExecutionType(&models.TTradeOrder{OrderType: legacyOrderTypeStopLoss, Price: 10}); got != int64(trade.OrderType_ORDER_TYPE_LIMIT) {
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
		Price:     100,
		Qty:       10,
	}
	opposites := []*models.TTradeOrder{
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 99, Qty: 4},
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 100, Qty: 6},
	}
	if !canFullyFillFromBook(order, opposites) {
		t.Fatal("order should be fully fillable across multiple levels")
	}

	largeFOK := []*models.TTradeOrder{
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_FOK), Price: 99, Qty: 20},
	}
	if canFullyFillFromBook(order, largeFOK) {
		t.Fatal("order should not partially fill an opposite FOK order")
	}

	withSkippedFOK := append(largeFOK, &models.TTradeOrder{
		Side:      int64(common.Side_SIDE_SELL),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Price:     100,
		Qty:       10,
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
		Price:       100,
		Qty:         10,
	}
	opposites := []*models.TTradeOrder{
		{Id: 2, Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_POST_ONLY), Price: 99, Qty: 10},
	}
	if canFullyFillFromBook(order, opposites) {
		t.Fatal("post-only liquidity that would take should not satisfy FOK")
	}
}

func TestCanFullyFillFromBookByAmount(t *testing.T) {
	order := &models.TTradeOrder{
		Side:      int64(common.Side_SIDE_BUY),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Price:     100,
		Amount:    10000,
	}
	opposites := []*models.TTradeOrder{
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 10, Qty: 5},
		{Side: int64(common.Side_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 20, Qty: 2.5},
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
		OrderId:  1,
		Price:    "10",
		Qty:      "2",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if fill.Amount != 2000 {
		t.Fatalf("computed fill amount = %v, want 2000", fill.Amount)
	}
}

func TestOrderBookKeyAndMember(t *testing.T) {
	order := &models.TTradeOrder{
		Id:         123,
		TenantId:   7,
		SymbolId:   4,
		MarketType: int64(trade.MarketType_MARKET_TYPE_USDT_CONTRACT),
		Side:       int64(common.Side_SIDE_BUY),
	}
	if got, want := orderBookKey(order), "trade:book:7:3:4:buy"; got != want {
		t.Fatalf("orderBookKey() = %q, want %q", got, want)
	}
	member := orderBookMember(order.Id)
	if got, err := orderBookMemberID(member); err != nil || got != order.Id {
		t.Fatalf("orderBookMemberID() = %d, err = %v, want %d", got, err, order.Id)
	}
}

func TestOrderBookScorePriority(t *testing.T) {
	marketBuy := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_MARKET), Side: int64(common.Side_SIDE_BUY)}
	highBuy := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Side: int64(common.Side_SIDE_BUY), Price: 101}
	lowBuy := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Side: int64(common.Side_SIDE_BUY), Price: 100}
	lowSell := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Side: int64(common.Side_SIDE_SELL), Price: 100}
	highSell := &models.TTradeOrder{OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Side: int64(common.Side_SIDE_SELL), Price: 101}

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
