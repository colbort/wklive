package logic

import (
	"testing"

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
			order: &models.TTradeOrder{Qty: 10, Amount: 100},
			want:  int64(trade.OrderStatus_ORDER_STATUS_PENDING),
		},
		{
			name:  "partial by qty",
			order: &models.TTradeOrder{Qty: 10, Amount: 100, FilledQty: 4, FilledAmount: 40},
			want:  int64(trade.OrderStatus_ORDER_STATUS_PART_FILLED),
		},
		{
			name:  "filled by qty",
			order: &models.TTradeOrder{Qty: 10, Amount: 100, FilledQty: 10, FilledAmount: 100},
			want:  int64(trade.OrderStatus_ORDER_STATUS_FILLED),
		},
		{
			name:  "filled by amount when qty target missing",
			order: &models.TTradeOrder{Amount: 100, FilledAmount: 100},
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

func TestRemainingMatchQty(t *testing.T) {
	got := remainingMatchQty(&models.TTradeOrder{Amount: 100, FilledAmount: 25}, 5)
	if got != 15 {
		t.Fatalf("remainingMatchQty() = %v, want 15", got)
	}

	got = remainingMatchQty(&models.TTradeOrder{Qty: 10, FilledQty: 4}, 5)
	if got != 6 {
		t.Fatalf("remainingMatchQty() = %v, want 6", got)
	}
}

func TestCanApplyOrderFill(t *testing.T) {
	qtyOrder := &models.TTradeOrder{Qty: 10, FilledQty: 4}
	if !canApplyOrderFill(qtyOrder, &models.TTradeFill{Qty: 6, Amount: 60}) {
		t.Fatal("remaining qty should be fillable")
	}
	if canApplyOrderFill(qtyOrder, &models.TTradeFill{Qty: 7, Amount: 70}) {
		t.Fatal("fill should not exceed remaining qty")
	}

	amountOrder := &models.TTradeOrder{Qty: 10, FilledQty: 4, Amount: 100, FilledAmount: 40}
	if !canApplyOrderFill(amountOrder, &models.TTradeFill{Qty: 3, Amount: 60}) {
		t.Fatal("remaining amount should be fillable")
	}
	if canApplyOrderFill(amountOrder, &models.TTradeFill{Qty: 3, Amount: 61}) {
		t.Fatal("fill should not exceed remaining amount")
	}
}

func TestFillMatchesOrder(t *testing.T) {
	order := &models.TTradeOrder{OrderNo: "TRD1", UserId: 1, SymbolId: 2, MarketType: 1, Side: int64(trade.TradeSide_TRADE_SIDE_BUY)}
	if !fillMatchesOrder(order, &models.TTradeFill{OrderNo: "TRD1", UserId: 1, SymbolId: 2, MarketType: 1, Side: int64(trade.TradeSide_TRADE_SIDE_BUY)}) {
		t.Fatal("matching fill metadata should pass")
	}
	if fillMatchesOrder(order, &models.TTradeFill{OrderNo: "TRD2", UserId: 1, SymbolId: 2, MarketType: 1, Side: int64(trade.TradeSide_TRADE_SIDE_BUY)}) {
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
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, OrderType: int64(trade.OrderType_ORDER_TYPE_TAKE_PROFIT), Side: int64(trade.TradeSide_TRADE_SIDE_SELL)},
			price: 101,
			want:  true,
		},
		{
			name:  "sell stop loss triggers downward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, OrderType: int64(trade.OrderType_ORDER_TYPE_STOP_LOSS), Side: int64(trade.TradeSide_TRADE_SIDE_SELL)},
			price: 99,
			want:  true,
		},
		{
			name:  "buy take profit triggers downward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, OrderType: int64(trade.OrderType_ORDER_TYPE_TAKE_PROFIT), Side: int64(trade.TradeSide_TRADE_SIDE_BUY)},
			price: 99,
			want:  true,
		},
		{
			name:  "buy stop loss triggers upward",
			order: models.TTradeOrder{Status: base.Status, TriggerPrice: base.TriggerPrice, OrderType: int64(trade.OrderType_ORDER_TYPE_STOP_LOSS), Side: int64(trade.TradeSide_TRADE_SIDE_BUY)},
			price: 101,
			want:  true,
		},
		{
			name:  "non waiting order does not trigger",
			order: models.TTradeOrder{Status: int64(trade.OrderStatus_ORDER_STATUS_PENDING), TriggerPrice: base.TriggerPrice, OrderType: int64(trade.OrderType_ORDER_TYPE_STOP_LOSS), Side: int64(trade.TradeSide_TRADE_SIDE_SELL)},
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
	if got := triggeredOrderType(&models.TTradeOrder{Price: 10}); got != int64(trade.OrderType_ORDER_TYPE_LIMIT) {
		t.Fatalf("triggeredOrderType() = %d, want LIMIT", got)
	}
	if got := triggeredOrderType(&models.TTradeOrder{}); got != int64(trade.OrderType_ORDER_TYPE_MARKET) {
		t.Fatalf("triggeredOrderType() = %d, want MARKET", got)
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
		Side:      int64(trade.TradeSide_TRADE_SIDE_BUY),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Price:     100,
		Qty:       10,
	}
	opposites := []*models.TTradeOrder{
		{Side: int64(trade.TradeSide_TRADE_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 99, Qty: 4},
		{Side: int64(trade.TradeSide_TRADE_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 100, Qty: 6},
	}
	if !canFullyFillFromBook(order, opposites) {
		t.Fatal("order should be fully fillable across multiple levels")
	}

	largeFOK := []*models.TTradeOrder{
		{Side: int64(trade.TradeSide_TRADE_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), TimeInForce: int64(trade.TimeInForce_TIME_IN_FORCE_FOK), Price: 99, Qty: 20},
	}
	if canFullyFillFromBook(order, largeFOK) {
		t.Fatal("order should not partially fill an opposite FOK order")
	}

	withSkippedFOK := append(largeFOK, &models.TTradeOrder{
		Side:      int64(trade.TradeSide_TRADE_SIDE_SELL),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Price:     100,
		Qty:       10,
	})
	if !canFullyFillFromBook(order, withSkippedFOK) {
		t.Fatal("order should skip incompatible FOK liquidity and fill from the next level")
	}
}

func TestCanFullyFillFromBookByAmount(t *testing.T) {
	order := &models.TTradeOrder{
		Side:      int64(trade.TradeSide_TRADE_SIDE_BUY),
		OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT),
		Price:     100,
		Amount:    100,
	}
	opposites := []*models.TTradeOrder{
		{Side: int64(trade.TradeSide_TRADE_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 10, Qty: 5},
		{Side: int64(trade.TradeSide_TRADE_SIDE_SELL), OrderType: int64(trade.OrderType_ORDER_TYPE_LIMIT), Price: 20, Qty: 2.5},
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
