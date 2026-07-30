package applogic

import (
	"context"
	"errors"
	"time"

	"wklive/common/generate"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// MatchFundedOrder admits a successfully funded order to matching.
// It is exported for the asset-instruction recovery task; callers must ensure
// the order is already in PENDING state.
func MatchFundedOrder(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TOptionOrder) error {
	if order == nil || order.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) {
		return nil
	}
	contract, err := svcCtx.OptionContractModel.FindOne(ctx, order.ContractId)
	if err != nil {
		return err
	}
	if contract.TenantId != order.TenantId ||
		contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) {
		return nil
	}
	logic := NewPlaceOrderLogic(ctx, svcCtx)
	if err := logic.matchOrder(contract, order); err != nil {
		return err
	}
	publishOptionOrderChanged(ctx, svcCtx, order)
	return nil
}

func (l *PlaceOrderLogic) matchOrder(contract *models.TOptionContract, order *models.TOptionOrder) error {
	if !order.Price.IsPositive() || !order.UnfilledQty.IsPositive() {
		return nil
	}

	changedOrders := make(map[int64]*models.TOptionOrder)
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		tradeModel := models.NewTOptionTradeModel(conn, l.svcCtx.Config.CacheRedis)
		matchSequenceModel := models.NewTOptionMatchSequenceModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		outboxModel := models.NewTOptionOutboxModel(conn, l.svcCtx.Config.CacheRedis)
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)

		incoming, err := orderModel.FindOne(ctx, order.Id)
		if err != nil {
			return err
		}
		if incoming.OrderType == int64(option.OrderType_ORDER_TYPE_POST_ONLY) ||
			incoming.OrderType == int64(option.OrderType_ORDER_TYPE_FOK) {
			candidates, err := orderModel.FindAllMatchableOrders(ctx, incoming.TenantId, incoming.ContractId, oppositeOrderSide(incoming.Side), incoming.UserId, incoming.AccountId, incoming.Price)
			if err != nil {
				return err
			}
			matchableQty := matchableOptionQty(incoming, candidates)
			if incoming.OrderType == int64(option.OrderType_ORDER_TYPE_POST_ONLY) {
				if matchableQty.IsPositive() {
					if err := cancelImmediateOrder(ctx, positionModel, instructionModel, incoming, "POST_ONLY_WOULD_TAKE", time.Now().Unix()); err != nil {
						return err
					}
					if err := orderModel.Update(ctx, incoming); err != nil {
						return err
					}
					changedOrders[incoming.Id] = incoming
				}
				*order = *incoming
				return nil
			}
			if matchableQty.LessThan(incoming.UnfilledQty) {
				if err := cancelImmediateOrder(ctx, positionModel, instructionModel, incoming, "FOK_NOT_FILLED", time.Now().Unix()); err != nil {
					return err
				}
				if err := orderModel.Update(ctx, incoming); err != nil {
					return err
				}
				changedOrders[incoming.Id] = incoming
				*order = *incoming
				return nil
			}
		}
		for incoming.UnfilledQty.IsPositive() {
			makers, err := orderModel.FindMatchableOrders(ctx, incoming.TenantId, incoming.ContractId, oppositeOrderSide(incoming.Side), incoming.UserId, incoming.AccountId, incoming.Price, 50)
			if err != nil {
				return err
			}
			if len(makers) == 0 {
				break
			}

			matched := false
			for _, maker := range makers {
				if !incoming.UnfilledQty.IsPositive() {
					break
				}
				if maker.Id == incoming.Id || !maker.UnfilledQty.IsPositive() {
					continue
				}
				if maker.UserId == incoming.UserId && maker.AccountId == incoming.AccountId {
					continue
				}

				tradeQty := decimal.Min(incoming.UnfilledQty, maker.UnfilledQty)
				if !tradeQty.IsPositive() {
					continue
				}
				tradePrice := maker.Price
				tradeNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "OT", "")
				if err != nil {
					return err
				}
				now := time.Now().Unix()
				matchSequence, err := matchSequenceModel.Next(ctx, incoming.TenantId, incoming.ContractId)
				if err != nil {
					return err
				}

				buyOrder := incoming
				sellOrder := maker
				if incoming.Side == int64(common.Side_SIDE_SELL) {
					buyOrder = maker
					sellOrder = incoming
				}

				trade := &models.TOptionTrade{
					TenantId:         incoming.TenantId,
					TradeNo:          tradeNo,
					ContractId:       incoming.ContractId,
					UnderlyingSymbol: incoming.UnderlyingSymbol,
					BuyOrderId:       buyOrder.Id,
					BuyOrderNo:       buyOrder.OrderNo,
					BuyUserId:        buyOrder.UserId,
					BuyAccountId:     buyOrder.AccountId,
					SellOrderId:      sellOrder.Id,
					SellOrderNo:      sellOrder.OrderNo,
					SellUserId:       sellOrder.UserId,
					SellAccountId:    sellOrder.AccountId,
					Price:            tradePrice,
					Qty:              tradeQty,
					Turnover:         optionTurnover(contract, tradePrice, tradeQty),
					FeeCoin:          contract.SettleCoin,
					MakerSide:        makerSide(maker.Side),
					MatchSequence:    matchSequence,
					TradeTime:        now,
					CreateTimes:      now,
				}
				trade.BuyFee, trade.SellFee = optionTradeFees(contract, trade.Turnover, trade.MakerSide)
				result, err := tradeModel.Insert(ctx, trade)
				if err != nil {
					return err
				}
				tradeId, err := result.LastInsertId()
				if err != nil {
					return err
				}
				trade.Id = tradeId
				sellerMargin := allocateSellerMargin(sellOrder, tradeQty, sellOrder.UnfilledQty)
				if sellerMargin.IsPositive() {
					if _, err := marginLotModel.Insert(ctx, &models.TOptionMarginLot{
						TenantId: sellOrder.TenantId, UserId: sellOrder.UserId, AccountId: sellOrder.AccountId,
						ContractId: sellOrder.ContractId, OrderId: sellOrder.Id, TradeId: trade.Id,
						FreezeBizNo: sellOrder.OrderNo, Quantity: tradeQty, RemainingQuantity: tradeQty,
						InitialMargin: sellerMargin, RemainingMargin: sellerMargin,
						Status:      int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
						CreateTimes: now, UpdateTimes: now,
					}); err != nil {
						return err
					}
				}
				if _, err := outboxModel.Insert(ctx, &models.TOptionOutbox{
					TenantId: incoming.TenantId, EventNo: trade.TradeNo + "-POSITION",
					EventType:  int64(option.OptionEventType_OPTION_EVENT_TYPE_TRADE_POSITION),
					ContractId: incoming.ContractId, TradeId: trade.Id, MatchSequence: trade.MatchSequence,
					Status:      int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
					CreateTimes: now, UpdateTimes: now,
				}); err != nil {
					return err
				}

				applyTradeToOrder(incoming, contract, tradePrice, tradeQty, now)
				applyTradeToOrder(maker, contract, tradePrice, tradeQty, now)
				buyOrder.Fee = buyOrder.Fee.Add(trade.BuyFee)
				sellOrder.Fee = sellOrder.Fee.Add(trade.SellFee)
				consumeBuyOrderReservation(buyOrder, trade.Turnover.Add(trade.BuyFee))
				if err := createTradeAssetInstructions(ctx, instructionModel, contract, trade, buyOrder, sellOrder, now); err != nil {
					return err
				}
				if err := orderModel.Update(ctx, maker); err != nil {
					return err
				}
				if err := orderModel.Update(ctx, incoming); err != nil {
					return err
				}
				incomingCopy := *incoming
				makerCopy := *maker
				changedOrders[incoming.Id] = &incomingCopy
				changedOrders[maker.Id] = &makerCopy
				if err := updateMarketLastTrade(ctx, marketModel, contract, tradePrice, now); err != nil {
					return err
				}
				matched = true
			}
			if !matched {
				break
			}
		}
		if incoming.UnfilledQty.IsPositive() && isImmediateOptionOrder(incoming.OrderType) {
			if err := cancelImmediateOrder(ctx, positionModel, instructionModel, incoming, "IMMEDIATE_REMAINDER_CANCELED", time.Now().Unix()); err != nil {
				return err
			}
			if err := orderModel.Update(ctx, incoming); err != nil {
				return err
			}
			incomingCopy := *incoming
			changedOrders[incoming.Id] = &incomingCopy
		}
		*order = *incoming
		return nil
	})
	if err != nil {
		return err
	}
	for _, changedOrder := range changedOrders {
		publishOptionOrderChanged(l.ctx, l.svcCtx, changedOrder)
	}
	return nil
}

func matchableOptionQty(incoming *models.TOptionOrder, candidates []*models.TOptionOrder) decimal.Decimal {
	total := decimal.Zero
	for _, candidate := range candidates {
		if candidate.Id == incoming.Id || !candidate.UnfilledQty.IsPositive() ||
			(candidate.UserId == incoming.UserId && candidate.AccountId == incoming.AccountId) {
			continue
		}
		total = total.Add(candidate.UnfilledQty)
	}
	return total
}

func isImmediateOptionOrder(orderType int64) bool {
	return orderType == int64(option.OrderType_ORDER_TYPE_MARKET) ||
		orderType == int64(option.OrderType_ORDER_TYPE_IOC) ||
		orderType == int64(option.OrderType_ORDER_TYPE_FOK)
}

func cancelImmediateOrder(ctx context.Context, positionModel models.TOptionPositionModel, instructionModel models.TOptionAssetInstructionModel, order *models.TOptionOrder, reason string, now int64) error {
	if !order.UnfilledQty.IsPositive() {
		return nil
	}
	if err := releaseClosePositionFrozenQty(ctx, positionModel, order, order.UnfilledQty, now); err != nil {
		return err
	}
	order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
	order.CancelReason = reason
	order.CancelTime = now
	order.UpdateTimes = now
	if order.MarginAmount.IsPositive() {
		order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELING)
		if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: order.TenantId, InstructionNo: order.OrderNo + "-IMMEDIATE-RELEASE",
			BizNo: order.OrderNo, OrderId: order.Id, UserId: order.UserId, AccountId: order.AccountId,
			Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
			TargetBizNo: order.OrderNo, Coin: order.FeeCoin, Amount: order.MarginAmount,
			StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		order.MarginAmount = decimal.Zero
	}
	return nil
}

func makerSide(side int64) int64 {
	if side == int64(common.Side_SIDE_BUY) {
		return int64(common.Side_SIDE_BUY)
	}
	if side == int64(common.Side_SIDE_SELL) {
		return int64(common.Side_SIDE_SELL)
	}
	return int64(common.Side_SIDE_UNKNOWN)
}

func updateMarketLastTrade(ctx context.Context, model models.TOptionMarketModel, contract *models.TOptionContract, price decimal.Decimal, now int64) error {
	market, err := model.FindOneByTenantIdContractId(ctx, contract.TenantId, contract.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if errors.Is(err, models.ErrNotFound) {
		_, err = model.Insert(ctx, &models.TOptionMarket{
			TenantId:   contract.TenantId,
			ContractId: contract.Id,
			MarkPrice:  price,
			LastPrice:  price,
			// A trade price is not an authoritative underlying quote. Keep the
			// quote timestamp empty so expiry settlement cannot mistake this
			// update for a fresh underlying-price snapshot.
			SnapshotTime:     0,
			PricingModel:     "trade",
			CreateTimes:      now,
			UpdateTimes:      now,
			TheoreticalPrice: price,
		})
		return err
	}

	market.LastPrice = price
	if !market.MarkPrice.IsPositive() {
		market.MarkPrice = price
	}
	if !market.TheoreticalPrice.IsPositive() {
		market.TheoreticalPrice = price
	}
	// SnapshotTime belongs to the underlying quote written by SyncMarketQuote.
	// Updating it here could make a stale settlement price appear fresh.
	market.UpdateTimes = now
	return model.Update(ctx, market)
}

func consumeBuyOrderReservation(order *models.TOptionOrder, turnover decimal.Decimal) {
	if order.Side != int64(common.Side_SIDE_BUY) || !turnover.IsPositive() {
		return
	}
	order.MarginAmount = decimal.Max(order.MarginAmount.Sub(turnover), decimal.Zero)
}

func allocateSellerMargin(order *models.TOptionOrder, fillQty, unfilledBefore decimal.Decimal) decimal.Decimal {
	if order == nil || order.Side != int64(common.Side_SIDE_SELL) ||
		order.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_OPEN) ||
		!order.MarginAmount.IsPositive() || !fillQty.IsPositive() || !unfilledBefore.IsPositive() {
		return decimal.Zero
	}
	allocated := order.MarginAmount
	if fillQty.LessThan(unfilledBefore) {
		allocated = order.MarginAmount.Mul(fillQty).Div(unfilledBefore).Round(16)
	}
	allocated = decimal.Min(allocated, order.MarginAmount)
	order.MarginAmount = decimal.Max(order.MarginAmount.Sub(allocated), decimal.Zero)
	return allocated
}

func optionTradeFees(contract *models.TOptionContract, turnover decimal.Decimal, makerSide int64) (buyFee, sellFee decimal.Decimal) {
	if contract == nil || !turnover.IsPositive() {
		return decimal.Zero, decimal.Zero
	}
	buyRate, sellRate := contract.TakerFeeRate, contract.MakerFeeRate
	if makerSide == int64(common.Side_SIDE_BUY) {
		buyRate, sellRate = contract.MakerFeeRate, contract.TakerFeeRate
	}
	// Asset and Option ledgers persist DECIMAL(32,16); round once at the
	// business boundary and reuse the exact amounts for every instruction.
	return turnover.Mul(buyRate).Round(16), turnover.Mul(sellRate).Round(16)
}

func createTradeAssetInstructions(ctx context.Context, model models.TOptionAssetInstructionModel, contract *models.TOptionContract, trade *models.TOptionTrade, buyOrder, sellOrder *models.TOptionOrder, now int64) error {
	if !trade.Turnover.IsPositive() {
		return nil
	}
	if err := validateOptionTradeAssetBalance(trade); err != nil {
		return err
	}
	if _, err := model.Insert(ctx, &models.TOptionAssetInstruction{
		TenantId: trade.TenantId, InstructionNo: trade.TradeNo + "-BUY-PREMIUM",
		BizNo: trade.TradeNo, OrderId: buyOrder.Id, TradeId: trade.Id,
		UserId: buyOrder.UserId, AccountId: buyOrder.AccountId,
		Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
		TargetBizNo: buyOrder.OrderNo, Coin: trade.FeeCoin, Amount: trade.Turnover.Add(trade.BuyFee),
		StepNo: 1, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
		CreateTimes:          now, UpdateTimes: now,
	}); err != nil {
		return err
	}
	sellerNet := trade.Turnover.Sub(trade.SellFee)
	if sellerNet.IsPositive() {
		if _, err := model.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: trade.TenantId, InstructionNo: trade.TradeNo + "-SELL-PREMIUM",
			BizNo: trade.TradeNo, OrderId: sellOrder.Id, TradeId: trade.Id,
			UserId: sellOrder.UserId, AccountId: sellOrder.AccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
			Coin:   trade.FeeCoin, Amount: sellerNet,
			StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return err
		}
	}
	totalFee := trade.BuyFee.Add(trade.SellFee)
	if totalFee.IsPositive() {
		if contract.FeeUserId <= 0 || contract.FeeAccountId <= 0 {
			return errors.New("option fee account is missing")
		}
		if _, err := model.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: trade.TenantId, InstructionNo: trade.TradeNo + "-PLATFORM-FEE",
			BizNo: trade.TradeNo, TradeId: trade.Id,
			UserId: contract.FeeUserId, AccountId: contract.FeeAccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
			Coin:   trade.FeeCoin, Amount: totalFee,
			StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return err
		}
	}
	if buyOrder.Status == int64(option.OrderStatus_ORDER_STATUS_FILLED) && buyOrder.MarginAmount.IsPositive() {
		if _, err := model.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: trade.TenantId, InstructionNo: buyOrder.OrderNo + "-RELEASE-REMAINDER",
			BizNo: buyOrder.OrderNo, OrderId: buyOrder.Id, TradeId: trade.Id,
			UserId: buyOrder.UserId, AccountId: buyOrder.AccountId,
			Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
			TargetBizNo: buyOrder.OrderNo, Coin: trade.FeeCoin, Amount: buyOrder.MarginAmount,
			StepNo: 3, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		buyOrder.MarginAmount = decimal.Zero
	}
	return nil
}

func validateOptionTradeAssetBalance(trade *models.TOptionTrade) error {
	if trade == nil || trade.Turnover.IsNegative() || trade.BuyFee.IsNegative() || trade.SellFee.IsNegative() ||
		trade.SellFee.GreaterThan(trade.Turnover) {
		return errors.New("invalid option trade asset amounts")
	}
	debit := trade.Turnover.Add(trade.BuyFee)
	credit := trade.Turnover.Sub(trade.SellFee).Add(trade.BuyFee).Add(trade.SellFee)
	if !debit.Equal(credit) {
		return errors.New("option trade asset instructions are not balanced")
	}
	return nil
}
