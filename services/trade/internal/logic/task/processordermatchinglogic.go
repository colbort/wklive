package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/domain/contractmath"
	"wklive/services/trade/internal/realtime"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	orderMatchKeyLimit  = int64(200)
	orderMatchBookDepth = int64(50)
	orderMatchMaxPerKey = 200
)

type ProcessOrderMatchingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type orderMatchPlan struct {
	BuyOrder  *models.TTradeOrder
	SellOrder *models.TTradeOrder
	Price     decimal.Decimal
	Qty       decimal.Decimal
	Amount    decimal.Decimal
}

func NewProcessOrderMatchingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessOrderMatchingLogic {
	return &ProcessOrderMatchingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订单撮合
func (l *ProcessOrderMatchingLogic) ProcessOrderMatching(in *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_order_matching", func(taskCtx context.Context) (*trade.TradeTaskResp, error) {
		l.ctx = taskCtx
		keys, err := l.svcCtx.TradeOrderModel.FindMatchKeys(l.ctx, in.GetTenantId(), helpers.MatchableOrderStatuses(), orderMatchKeyLimit)
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			if err := l.matchOrderBook(key); err != nil {
				return nil, err
			}
		}
		return helpers.OkTaskResp(), nil
	})
}

// ProcessOrder is the real-time entry point. Events for the same order book
// are serialized by a symbol-scoped distributed lock; different symbols can
// match concurrently.
func (l *ProcessOrderMatchingLogic) ProcessOrder(orderID int64) error {
	order, err := l.svcCtx.TradeOrderModel.FindOne(l.ctx, orderID)
	if err != nil {
		return err
	}
	if !helpers.IsMatchableOrderStatus(order.Status) {
		return nil
	}
	key := models.TradeOrderMatchKey{TenantId: order.TenantId, SymbolId: order.SymbolId, ProductType: order.ProductType}
	lockKey := fmt.Sprintf("trade:matching:%d:%d:%d", key.TenantId, key.ProductType, key.SymbolId)
	lockValue := fmt.Sprintf("%d", time.Now().UnixNano())
	if err := helpers.AcquireTaskLock(l.ctx, l.svcCtx.Redis, lockKey, lockValue); err != nil {
		if i18n.IsStatusError(err, i18n.SyncTaskAlreadyRunning) {
			return nil
		}
		return err
	}
	defer func() {
		if err := helpers.ReleaseTaskLock(context.Background(), l.svcCtx.Redis, lockKey, lockValue); err != nil {
			l.Errorf("release real-time matching lock failed, key=%s err=%v", lockKey, err)
		}
	}()
	return l.matchOrderBook(key)
}

func (l *ProcessOrderMatchingLogic) matchOrderBook(key models.TradeOrderMatchKey) error {
	for i := 0; i < orderMatchMaxPerKey; i++ {
		plan, err := l.findNextOrderMatch(key)
		if err != nil {
			return err
		}
		if plan == nil {
			return l.cancelResidualImmediateOrders(key)
		}
		matched, err := l.executeOrderMatch(key, plan)
		if err != nil {
			return err
		}
		if !matched {
			return l.cancelResidualImmediateOrders(key)
		}
	}
	return l.cancelResidualImmediateOrders(key)
}

func (l *ProcessOrderMatchingLogic) findNextOrderMatch(key models.TradeOrderMatchKey) (*orderMatchPlan, error) {
	buys, err := l.findOpenMatchOrdersFromBook(key, int64(common.Side_SIDE_BUY), orderMatchBookDepth)
	if err != nil {
		return nil, err
	}
	sells, err := l.findOpenMatchOrdersFromBook(key, int64(common.Side_SIDE_SELL), orderMatchBookDepth)
	if err != nil {
		return nil, err
	}

	return selectOrderMatchPlan(buys, sells), nil
}

func (l *ProcessOrderMatchingLogic) findOpenMatchOrdersFromBook(key models.TradeOrderMatchKey, side, limit int64) ([]*models.TTradeOrder, error) {
	orders, err := l.findOpenMatchOrdersFromCache(key, side, limit)
	if err != nil {
		l.Errorf("find match orders from redis book failed, tenantId=%d symbolId=%d productType=%d side=%d err=%v",
			key.TenantId, key.SymbolId, key.ProductType, side, err)
	} else if len(orders) > 0 {
		return orders, nil
	}

	orders, err = l.svcCtx.TradeOrderModel.FindOpenMatchOrders(
		l.ctx,
		key.TenantId,
		key.SymbolId,
		key.ProductType,
		side,
		helpers.MatchableOrderStatuses(),
		int64(trade.OrderType_ORDER_TYPE_MARKET),
		limit,
	)
	if err != nil {
		return nil, err
	}
	for _, order := range orders {
		if err := cacheOrderBookOrder(l.svcCtx, l.ctx, order); err != nil {
			l.Errorf("warm redis order book failed, orderId=%d err=%v", order.Id, err)
		}
	}
	return orders, nil
}

func (l *ProcessOrderMatchingLogic) findOpenMatchOrdersFromCache(key models.TradeOrderMatchKey, side, limit int64) ([]*models.TTradeOrder, error) {
	if l.svcCtx == nil || l.svcCtx.Redis == nil || limit <= 0 {
		return nil, nil
	}
	cacheKey := orderBookKeyBySide(key.TenantId, key.SymbolId, key.ProductType, side)
	scanLimit := limit * orderBookScanFactor
	members, err := l.svcCtx.Redis.ZrangeCtx(l.ctx, cacheKey, 0, scanLimit-1)
	if err != nil {
		return nil, err
	}
	orders := make([]*models.TTradeOrder, 0, limit)
	for _, member := range members {
		orderID, err := orderBookMemberID(member)
		if err != nil {
			continue
		}
		order, err := l.svcCtx.TradeOrderModel.FindOne(l.ctx, orderID)
		if errors.Is(err, models.ErrNotFound) {
			if removeErr := removeOrderBookMember(l.svcCtx, l.ctx, cacheKey, orderID); removeErr != nil {
				return nil, removeErr
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		if !sameMatchBook(order, key, side) || !isOrderBookOrder(order) {
			if removeErr := removeOrderBookMember(l.svcCtx, l.ctx, cacheKey, orderID); removeErr != nil {
				return nil, removeErr
			}
			continue
		}
		orders = append(orders, order)
		if int64(len(orders)) >= limit {
			break
		}
	}
	return orders, nil
}

func selectOrderMatchPlan(buys, sells []*models.TTradeOrder) *orderMatchPlan {
	for _, buy := range buys {
		for _, sell := range sells {
			plan := buildOrderMatchPlan(buy, sell)
			if plan == nil {
				continue
			}
			if isFOKOrder(buy) && !canFullyFillFromBook(buy, sells) {
				continue
			}
			if isFOKOrder(sell) && !canFullyFillFromBook(sell, buys) {
				continue
			}
			return plan
		}
	}
	return nil
}

func (l *ProcessOrderMatchingLogic) executeOrderMatch(key models.TradeOrderMatchKey, plan *orderMatchPlan) (bool, error) {
	if plan == nil || plan.BuyOrder == nil || plan.SellOrder == nil {
		return false, nil
	}
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, key.SymbolId)
	if err != nil {
		return false, err
	}
	makerFeeRate, takerFeeRate, contractConfig, err := l.matchFeeRates(key)
	if err != nil {
		return false, err
	}

	now := utils.NowMillis()
	matched := false
	matchedOrderIDs := make(map[int64]struct{})
	var createdFillEvents []realtime.Event
	err = helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		fillModel := models.NewTTradeFillModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis)
		contractOrderModel := models.NewTTradeOrderContractModel(conn, l.svcCtx.Config.CacheRedis)

		buy, err := orderModel.FindOneForUpdate(ctx, plan.BuyOrder.Id)
		if err != nil {
			return err
		}
		sell, err := orderModel.FindOneForUpdate(ctx, plan.SellOrder.Id)
		if err != nil {
			return err
		}
		lockedPlans, ok, err := l.normalizeLockedMatchPlans(ctx, orderModel, key, buy, sell)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		for _, lockedPlan := range lockedPlans {
			if lockedPlan == nil || !sameMatchBook(lockedPlan.BuyOrder, key, int64(common.Side_SIDE_BUY)) || !sameMatchBook(lockedPlan.SellOrder, key, int64(common.Side_SIDE_SELL)) {
				return nil
			}
			matchNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "MAT", "")
			if err != nil {
				return err
			}
			buyFillNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "FIL", "")
			if err != nil {
				return err
			}
			sellFillNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "FIL", "")
			if err != nil {
				return err
			}

			buyLiquidity := liquidityTypeForOrder(lockedPlan.BuyOrder, lockedPlan.SellOrder)
			sellLiquidity := liquidityTypeForOrder(lockedPlan.SellOrder, lockedPlan.BuyOrder)
			var buyFee, sellFee decimal.Decimal
			if symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
				values, err := contractmath.CalculateTradeValues(symbol.ContractValueType, lockedPlan.Qty, contractConfig.ContractSize, lockedPlan.Price)
				if err != nil {
					return err
				}
				lockedPlan.Amount = values.QuoteNotional
				buyFee = contractmath.CalculateFee(values, feeRateByLiquidity(buyLiquidity, makerFeeRate, takerFeeRate))
				sellFee = contractmath.CalculateFee(values, feeRateByLiquidity(sellLiquidity, makerFeeRate, takerFeeRate))
			} else {
				buyFee = matchFeeAmount(lockedPlan.Amount, buyLiquidity, makerFeeRate, takerFeeRate)
				sellFee = matchFeeAmount(lockedPlan.Amount, sellLiquidity, makerFeeRate, takerFeeRate)
			}

			buyFill, buyOrder, err := recordOrderFillWithModels(ctx, fillModel, orderModel, buildMatchFill(lockedPlan.BuyOrder, matchNo, buyFillNo, buyLiquidity, lockedPlan, buyFee, feeAssetForOrder(lockedPlan.BuyOrder, symbol), now), now)
			if err != nil {
				return err
			}
			sellFill, sellOrder, err := recordOrderFillWithModels(ctx, fillModel, orderModel, buildMatchFill(lockedPlan.SellOrder, matchNo, sellFillNo, sellLiquidity, lockedPlan, sellFee, feeAssetForOrder(lockedPlan.SellOrder, symbol), now), now)
			if err != nil {
				return err
			}
			if err := createMatchSettlementRecords(ctx, instructionModel, eventModel, contractOrderModel, symbol, buyOrder, buyFill, now); err != nil {
				return err
			}
			if err := createMatchSettlementRecords(ctx, instructionModel, eventModel, contractOrderModel, symbol, sellOrder, sellFill, now); err != nil {
				return err
			}
			createdFillEvents = append(createdFillEvents,
				realtime.Event{EventNo: derivedTradeBizNo(buyFill.FillNo, "FILL"), Type: realtime.EventFillCreated, TenantID: buyFill.TenantId, BizID: buyFill.FillNo, OrderID: buyFill.OrderId, FillID: buyFill.Id},
				realtime.Event{EventNo: derivedTradeBizNo(sellFill.FillNo, "FILL"), Type: realtime.EventFillCreated, TenantID: sellFill.TenantId, BizID: sellFill.FillNo, OrderID: sellFill.OrderId, FillID: sellFill.Id},
			)
			if symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
				createdFillEvents = append(createdFillEvents,
					realtime.Event{EventNo: derivedTradeBizNo(buyFill.FillNo, "POSITION"), Type: realtime.EventPositionFill, TenantID: buyFill.TenantId, BizID: buyFill.FillNo, OrderID: buyFill.OrderId, FillID: buyFill.Id},
					realtime.Event{EventNo: derivedTradeBizNo(sellFill.FillNo, "POSITION"), Type: realtime.EventPositionFill, TenantID: sellFill.TenantId, BizID: sellFill.FillNo, OrderID: sellFill.OrderId, FillID: sellFill.Id},
				)
			}
			matchedOrderIDs[lockedPlan.BuyOrder.Id] = struct{}{}
			matchedOrderIDs[lockedPlan.SellOrder.Id] = struct{}{}
			matched = true
		}
		return nil
	})
	if err != nil || !matched {
		return matched, err
	}
	for _, event := range createdFillEvents {
		if publishErr := helpers.PublishTradeOutboxEvent(l.ctx, l.svcCtx, event); publishErr != nil {
			l.Errorf("publish real-time fill event failed, eventNo=%s fillId=%d err=%v", event.EventNo, event.FillID, publishErr)
		}
	}
	for orderID := range matchedOrderIDs {
		if syncErr := syncOrderBookCacheByID(l.svcCtx, l.ctx, orderID); syncErr != nil {
			l.Errorf("sync redis order book after match failed, orderId=%d err=%v", orderID, syncErr)
		}
	}
	return matched, nil
}

func (l *ProcessOrderMatchingLogic) normalizeLockedMatchPlans(ctx context.Context, orderModel models.TTradeOrderModel, key models.TradeOrderMatchKey, buy, sell *models.TTradeOrder) ([]*orderMatchPlan, bool, error) {
	lockedPlan, ok := normalizeLockedMatchPlan(buy, sell)
	if !ok {
		return nil, false, nil
	}
	switch {
	case isFOKOrder(buy):
		return l.buildFOKMatchPlans(ctx, orderModel, key, buy)
	case isFOKOrder(sell):
		return l.buildFOKMatchPlans(ctx, orderModel, key, sell)
	default:
		return []*orderMatchPlan{lockedPlan}, true, nil
	}
}

func normalizeLockedMatchPlan(buy, sell *models.TTradeOrder) (*orderMatchPlan, bool) {
	if buy == nil || sell == nil || !helpers.IsMatchableOrderStatus(buy.Status) || !helpers.IsMatchableOrderStatus(sell.Status) {
		return nil, false
	}
	plan := buildOrderMatchPlan(buy, sell)
	if plan == nil {
		return nil, false
	}
	return plan, true
}

func buildOrderMatchPlan(buy, sell *models.TTradeOrder) *orderMatchPlan {
	if buy == nil || sell == nil {
		return nil
	}
	price, ok := matchExecutionPrice(buy, sell)
	if !ok {
		return nil
	}
	if postOnlyWouldTake(buy, sell) {
		return nil
	}
	buyQty := remainingMatchQty(buy, price)
	sellQty := remainingMatchQty(sell, price)
	qty := decimal.Min(buyQty, sellQty)
	if !qty.IsPositive() {
		return nil
	}
	amount := helpers.TradeMinorAmountAtPrice(price, qty)
	amount = clampMatchAmountToOrderRemainder(buy, sell, price, qty, amount)
	if !amount.IsPositive() {
		return nil
	}
	return &orderMatchPlan{
		BuyOrder:  buy,
		SellOrder: sell,
		Price:     price,
		Qty:       qty,
		Amount:    amount,
	}
}

func (l *ProcessOrderMatchingLogic) buildFOKMatchPlans(ctx context.Context, orderModel models.TTradeOrderModel, key models.TradeOrderMatchKey, focal *models.TTradeOrder) ([]*orderMatchPlan, bool, error) {
	if focal == nil || !helpers.IsMatchableOrderStatus(focal.Status) {
		return nil, false, nil
	}
	oppositeSide := int64(common.Side_SIDE_SELL)
	if focal.Side == int64(common.Side_SIDE_SELL) {
		oppositeSide = int64(common.Side_SIDE_BUY)
	}
	opposites, err := orderModel.FindOpenMatchOrders(ctx, key.TenantId, key.SymbolId, key.ProductType, oppositeSide, helpers.MatchableOrderStatuses(), int64(trade.OrderType_ORDER_TYPE_MARKET), orderMatchBookDepth)
	if err != nil {
		return nil, false, err
	}

	need := newOrderFillNeed(focal)
	plans := make([]*orderMatchPlan, 0, len(opposites))
	for _, item := range opposites {
		if need.filled() {
			break
		}
		opposite, err := orderModel.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return nil, false, err
		}
		if !sameMatchBook(opposite, key, oppositeSide) || !helpers.IsMatchableOrderStatus(opposite.Status) {
			continue
		}
		buy, sell := matchSides(focal, opposite)
		price, ok := matchExecutionPrice(buy, sell)
		if !ok || postOnlyWouldTake(buy, sell) {
			continue
		}

		focalQty := need.matchQty(price)
		oppositeQty := remainingMatchQty(opposite, price)
		if !focalQty.IsPositive() {
			break
		}
		if !oppositeQty.IsPositive() {
			continue
		}
		if isFOKOrder(opposite) && oppositeQty.GreaterThan(focalQty) {
			continue
		}

		qty := decimal.Min(focalQty, oppositeQty)
		amount := helpers.TradeMinorAmountAtPrice(price, qty)
		amount = clampMatchAmountToOrderRemainder(buy, sell, price, qty, amount)
		if !qty.IsPositive() || !amount.IsPositive() {
			continue
		}
		plans = append(plans, &orderMatchPlan{
			BuyOrder:  buy,
			SellOrder: sell,
			Price:     price,
			Qty:       qty,
			Amount:    amount,
		})
		need.consume(qty, amount)
	}
	if !need.filled() {
		return nil, false, nil
	}
	return plans, len(plans) > 0, nil
}

// clampMatchAmountToOrderRemainder removes division/multiplication dust for
// amount-based orders. When an amount order is the side limiting this match,
// its calculated qty is remainingAmount/price; multiplying it back by price
// may be microscopically smaller because decimal division has finite
// precision. Persisting that dust makes a fully consumed market order look
// partially filled and sends it into the residual-cancel path.
func clampMatchAmountToOrderRemainder(buy, sell *models.TTradeOrder, price, qty, amount decimal.Decimal) decimal.Decimal {
	for _, order := range []*models.TTradeOrder{buy, sell} {
		if order == nil || order.Qty.IsPositive() || !order.Amount.IsPositive() {
			continue
		}
		remaining := decimal.Max(order.Amount.Sub(order.FilledAmount), decimal.Zero)
		if !remaining.IsPositive() {
			continue
		}
		limitingQty := helpers.TradeQtyFromMinorAmount(remaining, price)
		if qty.Equal(limitingQty) {
			return remaining
		}
	}
	return amount
}

func sameMatchBook(order *models.TTradeOrder, key models.TradeOrderMatchKey, side int64) bool {
	return order != nil &&
		order.TenantId == key.TenantId &&
		order.SymbolId == key.SymbolId &&
		order.ProductType == key.ProductType &&
		order.Side == side
}

func matchSides(left, right *models.TTradeOrder) (*models.TTradeOrder, *models.TTradeOrder) {
	if left.Side == int64(common.Side_SIDE_BUY) {
		return left, right
	}
	return right, left
}

func buildMatchFill(order *models.TTradeOrder, matchNo, fillNo string, liquidity trade.LiquidityType, plan *orderMatchPlan, fee decimal.Decimal, feeAsset string, now int64) *trade.TradeFill {
	return &trade.TradeFill{
		TenantId:      order.TenantId,
		FillNo:        fillNo,
		MatchNo:       matchNo,
		OrderId:       order.Id,
		OrderNo:       order.OrderNo,
		UserId:        order.UserId,
		SymbolId:      order.SymbolId,
		ProductType:   common.ProductType(order.ProductType),
		Side:          common.Side(order.Side),
		PositionSide:  trade.PositionSide(order.PositionSide),
		Price:         conv.FloatString(plan.Price),
		Qty:           conv.FloatString(plan.Qty),
		Amount:        conv.FloatString(plan.Amount),
		Fee:           conv.FloatString(fee),
		FeeAsset:      feeAsset,
		LiquidityType: liquidity,
		RealizedPnl:   "0",
		MatchTime:     now,
		CreateTimes:   now,
	}
}

func matchExecutionPrice(buy, sell *models.TTradeOrder) (decimal.Decimal, bool) {
	buyMarket := buy.OrderType == int64(trade.OrderType_ORDER_TYPE_MARKET)
	sellMarket := sell.OrderType == int64(trade.OrderType_ORDER_TYPE_MARKET)
	switch {
	case buyMarket && sellMarket:
		return decimal.Zero, false
	case buyMarket:
		return sell.Price, sell.Price.IsPositive()
	case sellMarket:
		return buy.Price, buy.Price.IsPositive()
	case buy.Price.LessThan(sell.Price):
		return decimal.Zero, false
	case buy.Id < sell.Id:
		return buy.Price, buy.Price.IsPositive()
	default:
		return sell.Price, sell.Price.IsPositive()
	}
}

func remainingMatchQty(order *models.TTradeOrder, price decimal.Decimal) decimal.Decimal {
	if order.Qty.IsPositive() {
		return decimal.Max(order.Qty.Sub(order.FilledQty), decimal.Zero)
	}
	if order.Amount.IsPositive() && price.IsPositive() {
		return helpers.TradeQtyFromMinorAmount(decimal.Max(order.Amount.Sub(order.FilledAmount), decimal.Zero), price)
	}
	return decimal.Zero
}

type orderFillNeed struct {
	byQty           bool
	remainingQty    decimal.Decimal
	remainingAmount decimal.Decimal
}

func newOrderFillNeed(order *models.TTradeOrder) orderFillNeed {
	if order == nil {
		return orderFillNeed{}
	}
	if order.Qty.IsPositive() {
		return orderFillNeed{byQty: true, remainingQty: decimal.Max(order.Qty.Sub(order.FilledQty), decimal.Zero)}
	}
	return orderFillNeed{remainingAmount: decimal.Max(order.Amount.Sub(order.FilledAmount), decimal.Zero)}
}

func (n orderFillNeed) matchQty(price decimal.Decimal) decimal.Decimal {
	if n.byQty {
		return n.remainingQty
	}
	return helpers.TradeQtyFromMinorAmount(n.remainingAmount, price)
}

func (n *orderFillNeed) consume(qty, amount decimal.Decimal) {
	if n.byQty {
		n.remainingQty = decimal.Max(n.remainingQty.Sub(qty), decimal.Zero)
		return
	}
	n.remainingAmount = decimal.Max(n.remainingAmount.Sub(amount), decimal.Zero)
}

func (n orderFillNeed) filled() bool {
	if n.byQty {
		return !n.remainingQty.IsPositive()
	}
	return !n.remainingAmount.IsPositive()
}

func isFOKOrder(order *models.TTradeOrder) bool {
	return order != nil && order.TimeInForce == int64(trade.TimeInForce_TIME_IN_FORCE_FOK)
}

func isPostOnlyOrder(order *models.TTradeOrder) bool {
	return order != nil && order.TimeInForce == int64(trade.TimeInForce_TIME_IN_FORCE_POST_ONLY)
}

func canFullyFillFromBook(order *models.TTradeOrder, opposites []*models.TTradeOrder) bool {
	if order == nil {
		return false
	}
	need := newOrderFillNeed(order)
	for _, opposite := range opposites {
		var (
			price decimal.Decimal
			ok    bool
		)
		if order.Side == int64(common.Side_SIDE_BUY) {
			price, ok = matchExecutionPrice(order, opposite)
		} else {
			price, ok = matchExecutionPrice(opposite, order)
		}
		buy, sell := matchSides(order, opposite)
		if !ok || postOnlyWouldTake(buy, sell) {
			continue
		}
		focalQty := need.matchQty(price)
		oppositeQty := remainingMatchQty(opposite, price)
		if !focalQty.IsPositive() {
			return true
		}
		if !oppositeQty.IsPositive() {
			continue
		}
		if isFOKOrder(opposite) && oppositeQty.GreaterThan(focalQty) {
			continue
		}
		qty := decimal.Min(focalQty, oppositeQty)
		need.consume(qty, helpers.TradeMinorAmountAtPrice(price, qty))
		if need.filled() {
			return true
		}
	}
	return false
}

func postOnlyWouldTake(buy, sell *models.TTradeOrder) bool {
	if isPostOnlyOrder(buy) && liquidityTypeForOrder(buy, sell) == trade.LiquidityType_LIQUIDITY_TYPE_TAKER {
		return true
	}
	if isPostOnlyOrder(sell) && liquidityTypeForOrder(sell, buy) == trade.LiquidityType_LIQUIDITY_TYPE_TAKER {
		return true
	}
	return false
}

func liquidityTypeForOrder(order, peer *models.TTradeOrder) trade.LiquidityType {
	if order.Id < peer.Id {
		return trade.LiquidityType_LIQUIDITY_TYPE_MAKER
	}
	return trade.LiquidityType_LIQUIDITY_TYPE_TAKER
}

func matchFeeAmount(amount decimal.Decimal, liquidity trade.LiquidityType, makerFeeRate, takerFeeRate decimal.Decimal) decimal.Decimal {
	return amount.Mul(feeRateByLiquidity(liquidity, makerFeeRate, takerFeeRate))
}

func feeRateByLiquidity(liquidity trade.LiquidityType, makerFeeRate, takerFeeRate decimal.Decimal) decimal.Decimal {
	if liquidity == trade.LiquidityType_LIQUIDITY_TYPE_MAKER {
		return makerFeeRate
	}
	return takerFeeRate
}

func feeAssetForOrder(order *models.TTradeOrder, symbol *models.TTradeSymbol) string {
	if order != nil && order.ProductType == int64(common.ProductType_PRODUCT_TYPE_SPOT) {
		// Spot fees are denominated in quote asset in the current reservation
		// model, so they are always backed by the same frozen asset.
		return symbol.QuoteAsset
	}
	if order.FeeAsset != "" {
		return order.FeeAsset
	}
	return helpers.MarginAssetForSymbol(symbol)
}

func (l *ProcessOrderMatchingLogic) matchFeeRates(key models.TradeOrderMatchKey) (decimal.Decimal, decimal.Decimal, *models.TTradeSymbolContract, error) {
	if key.ProductType == int64(common.ProductType_PRODUCT_TYPE_SPOT) {
		spot, err := l.svcCtx.TradeSymbolSpotModel.FindOneByTenantIdSymbolId(l.ctx, key.TenantId, key.SymbolId)
		if err != nil {
			return decimal.Zero, decimal.Zero, nil, err
		}
		return spot.MakerFeeRate, spot.TakerFeeRate, nil, nil
	}
	contract, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, key.TenantId, key.SymbolId)
	if err != nil {
		return decimal.Zero, decimal.Zero, nil, err
	}
	return contract.MakerFeeRate, contract.TakerFeeRate, contract, nil
}

func (l *ProcessOrderMatchingLogic) cancelResidualImmediateOrders(key models.TradeOrderMatchKey) error {
	buys, err := l.svcCtx.TradeOrderModel.FindOpenMatchOrders(l.ctx, key.TenantId, key.SymbolId, key.ProductType, int64(common.Side_SIDE_BUY), helpers.MatchableOrderStatuses(), int64(trade.OrderType_ORDER_TYPE_MARKET), orderMatchBookDepth)
	if err != nil {
		return err
	}
	sells, err := l.svcCtx.TradeOrderModel.FindOpenMatchOrders(l.ctx, key.TenantId, key.SymbolId, key.ProductType, int64(common.Side_SIDE_SELL), helpers.MatchableOrderStatuses(), int64(trade.OrderType_ORDER_TYPE_MARKET), orderMatchBookDepth)
	if err != nil {
		return err
	}
	for _, order := range append(buys, sells...) {
		reason := residualCancelReason(order, buys, sells)
		if reason == "" {
			continue
		}
		canceledOrder, err := l.cancelOpenOrderNow(order.Id, reason)
		if err != nil {
			return err
		}
		if canceledOrder != nil {
			if err := unfreezeRemainingOrderAsset(l.svcCtx, l.ctx, canceledOrder, "trade matching residual unfreeze"); err != nil {
				return err
			}
			if err := removeOrderBookOrder(l.svcCtx, l.ctx, canceledOrder); err != nil {
				l.Errorf("remove canceled residual order from cache failed, orderId=%d err=%v", canceledOrder.Id, err)
			}
		}
	}
	return nil
}

func residualCancelReason(order *models.TTradeOrder, buys, sells []*models.TTradeOrder) string {
	if order.OrderType == int64(trade.OrderType_ORDER_TYPE_MARKET) {
		if order.FilledQty.IsPositive() || order.FilledAmount.IsPositive() {
			return "canceled: market order residual after partial fill"
		}
		return "canceled: market order has no executable liquidity"
	}
	switch trade.TimeInForce(order.TimeInForce) {
	case trade.TimeInForce_TIME_IN_FORCE_IOC:
		return "canceled: IOC residual"
	case trade.TimeInForce_TIME_IN_FORCE_FOK:
		return "canceled: FOK cannot be fully filled"
	case trade.TimeInForce_TIME_IN_FORCE_POST_ONLY:
		if postOnlyWouldTakeTop(order, buys, sells) {
			return "canceled: post-only order would take liquidity"
		}
	}
	return ""
}

func postOnlyWouldTakeTop(order *models.TTradeOrder, buys, sells []*models.TTradeOrder) bool {
	opposites := sells
	if order.Side == int64(common.Side_SIDE_SELL) {
		opposites = buys
	}
	for _, opposite := range opposites {
		var (
			price decimal.Decimal
			ok    bool
		)
		if order.Side == int64(common.Side_SIDE_BUY) {
			price, ok = matchExecutionPrice(order, opposite)
			if ok && price.IsPositive() && liquidityTypeForOrder(order, opposite) == trade.LiquidityType_LIQUIDITY_TYPE_TAKER {
				return true
			}
			continue
		}
		price, ok = matchExecutionPrice(opposite, order)
		if ok && price.IsPositive() && liquidityTypeForOrder(order, opposite) == trade.LiquidityType_LIQUIDITY_TYPE_TAKER {
			return true
		}
	}
	return false
}

func (l *ProcessOrderMatchingLogic) cancelOpenOrderNow(orderID int64, reason string) (*models.TTradeOrder, error) {
	now := utils.NowMillis()
	var canceledOrder *models.TTradeOrder
	err := helpers.TransactWithDeadlockRetry(l.ctx, l.svcCtx.DB, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTTradeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		order, err := orderModel.FindOneForUpdate(ctx, orderID)
		if err != nil {
			return err
		}
		if !helpers.IsMatchableOrderStatus(order.Status) {
			return nil
		}
		order.Status = int64(trade.OrderStatus_ORDER_STATUS_CANCELING)
		order.CanceledQty = decimalMaxZero(order.Qty.Sub(order.FilledQty))
		order.CancelReason = reason
		order.Version++
		order.UpdateTimes = now
		if err := orderModel.Update(ctx, order); err != nil {
			return err
		}
		canceledOrder = order
		return nil
	})
	if err != nil || canceledOrder == nil {
		return canceledOrder, err
	}
	return canceledOrder, nil
}
