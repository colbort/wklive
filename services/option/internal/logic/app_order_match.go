package logic

import (
	"context"
	"errors"
	"time"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func (l *AppPlaceOrderLogic) matchOrder(contract *models.TOptionContract, order *models.TOptionOrder) error {
	if order.OrderType == int64(option.OrderType_ORDER_TYPE_POST_ONLY) {
		return nil
	}
	if order.Price <= 0 || order.UnfilledQty <= optionFloatEpsilon {
		return nil
	}

	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		tradeModel := models.NewTOptionTradeModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis)
		accountModel := models.NewTOptionAccountModel(conn, l.svcCtx.Config.CacheRedis)
		billModel := models.NewTOptionBillModel(conn, l.svcCtx.Config.CacheRedis)

		incoming, err := orderModel.FindOne(ctx, order.Id)
		if err != nil {
			return err
		}
		for incoming.UnfilledQty > optionFloatEpsilon {
			makers, err := orderModel.FindMatchableOrders(ctx, incoming.TenantId, incoming.ContractId, oppositeOrderSide(incoming.Side), incoming.Price, 50)
			if err != nil {
				return err
			}
			if len(makers) == 0 {
				break
			}

			matched := false
			for _, maker := range makers {
				if incoming.UnfilledQty <= optionFloatEpsilon {
					break
				}
				if maker.Id == incoming.Id || maker.UnfilledQty <= optionFloatEpsilon {
					continue
				}
				if maker.UserId == incoming.UserId && maker.AccountId == incoming.AccountId {
					continue
				}

				tradeQty := minFloat64(incoming.UnfilledQty, maker.UnfilledQty)
				if tradeQty <= optionFloatEpsilon {
					continue
				}
				tradePrice := maker.Price
				tradeNo, err := l.svcCtx.GenerateBizNo(ctx, "OT")
				if err != nil {
					return err
				}
				now := time.Now().Unix()

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
					TradeTime:        now,
					CreateTimes:      now,
				}
				result, err := tradeModel.Insert(ctx, trade)
				if err != nil {
					return err
				}
				tradeId, err := result.LastInsertId()
				if err != nil {
					return err
				}
				trade.Id = tradeId

				applyTradeToOrder(incoming, contract, tradePrice, tradeQty, now)
				applyTradeToOrder(maker, contract, tradePrice, tradeQty, now)
				if err := updatePositionByFilledOrder(ctx, positionModel, contract, buyOrder, tradePrice, tradeQty, now); err != nil {
					return err
				}
				if err := updatePositionByFilledOrder(ctx, positionModel, contract, sellOrder, tradePrice, tradeQty, now); err != nil {
					return err
				}
				if err := orderModel.Update(ctx, maker); err != nil {
					return err
				}
				if err := orderModel.Update(ctx, incoming); err != nil {
					return err
				}
				if err := updateMarketLastTrade(ctx, marketModel, contract, tradePrice, now); err != nil {
					return err
				}
				if err := applyTradeAccountBills(ctx, accountModel, billModel, trade, now); err != nil {
					return err
				}

				matched = true
			}
			if !matched {
				break
			}
		}
		*order = *incoming
		return nil
	})
	if err != nil {
		return err
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

func updateMarketLastTrade(ctx context.Context, model models.TOptionMarketModel, contract *models.TOptionContract, price float64, now int64) error {
	market, err := model.FindOneByTenantIdContractId(ctx, contract.TenantId, contract.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if errors.Is(err, models.ErrNotFound) {
		_, err = model.Insert(ctx, &models.TOptionMarket{
			TenantId:         contract.TenantId,
			ContractId:       contract.Id,
			MarkPrice:        price,
			LastPrice:        price,
			SnapshotTime:     now,
			PricingModel:     "trade",
			CreateTimes:      now,
			UpdateTimes:      now,
			TheoreticalPrice: price,
		})
		return err
	}

	market.LastPrice = price
	if market.MarkPrice <= 0 {
		market.MarkPrice = price
	}
	if market.TheoreticalPrice <= 0 {
		market.TheoreticalPrice = price
	}
	market.SnapshotTime = now
	market.UpdateTimes = now
	return model.Update(ctx, market)
}

func applyTradeAccountBills(ctx context.Context, accountModel models.TOptionAccountModel, billModel models.TOptionBillModel, trade *models.TOptionTrade, now int64) error {
	if err := applyOptionAccountDelta(ctx, accountModel, billModel, trade.TenantId, trade.BuyUserId, trade.BuyAccountId, trade.FeeCoin, -trade.Turnover, int64(option.BillRefType_BILL_REF_TYPE_TRADE), trade.Id, trade.TradeNo+"-BUY", "option trade premium paid", false, now); err != nil {
		return err
	}
	return applyOptionAccountDelta(ctx, accountModel, billModel, trade.TenantId, trade.SellUserId, trade.SellAccountId, trade.FeeCoin, trade.Turnover, int64(option.BillRefType_BILL_REF_TYPE_TRADE), trade.Id, trade.TradeNo+"-SELL", "option trade premium received", false, now)
}
