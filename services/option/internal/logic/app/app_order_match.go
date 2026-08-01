package applogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/generate"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/observability"
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
	if err := rejectComboChildFromSimpleMatcher(order, "funded_entry"); err != nil {
		return err
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
	if err := rejectComboChildFromSimpleMatcher(order, "match_entry"); err != nil {
		return err
	}
	if !order.Price.IsPositive() || !order.UnfilledQty.IsPositive() {
		return nil
	}

	changedOrders := make(map[int64]*models.TOptionOrder)
	type mmpGroupKey struct {
		tenantID, userID, contractID int64
		groupCode                    string
	}
	triggeredMMPGroups := make(map[mmpGroupKey]struct{})
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		tradeModel := models.NewTOptionTradeModel(conn, l.svcCtx.Config.CacheRedis)
		matchSequenceModel := models.NewTOptionMatchSequenceModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		outboxModel := models.NewTOptionOutboxModel(conn, l.svcCtx.Config.CacheRedis)
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		controlEventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		userControlModel := models.NewTOptionUserTradingControlModel(conn, l.svcCtx.Config.CacheRedis)

		lockedContract, err := contractModel.FindOneForUpdate(ctx, order.ContractId)
		if err != nil {
			return err
		}
		lockedMarket, marketErr := marketModel.FindOneByTenantIdContractIdForUpdate(
			ctx, order.TenantId, order.ContractId,
		)
		if marketErr != nil && !errors.Is(marketErr, models.ErrNotFound) {
			return marketErr
		}
		incomingControl, err := userControlModel.EnsureForUpdate(
			ctx, order.TenantId, order.UserId, time.Now().Unix(),
		)
		if err != nil {
			return err
		}
		incoming, err := orderModel.FindOneForUpdate(ctx, order.Id)
		if err != nil {
			return err
		}
		if err := rejectComboChildFromSimpleMatcher(incoming, "locked_match"); err != nil {
			return err
		}
		if lockedContract.Id != incoming.ContractId ||
			incomingControl.TenantId != incoming.TenantId ||
			incomingControl.UserId != incoming.UserId {
			return errors.New("option matching lock scope does not match incoming order")
		}
		switch option.OrderStatus(incoming.Status) {
		case option.OrderStatus_ORDER_STATUS_PENDING,
			option.OrderStatus_ORDER_STATUS_PART_FILLED:
		default:
			*order = *incoming
			return nil
		}
		now := time.Now().Unix()
		if errors.Is(marketErr, models.ErrNotFound) ||
			!logichelpers.IsMarkFresh(lockedMarket, now, 30) ||
			!lockedMarket.MarkPrice.IsPositive() {
			if err := cancelImmediateOrder(
				ctx, positionModel, instructionModel, incoming, controlReasonStaleMark, now,
			); err != nil {
				return err
			}
			if err := orderModel.Update(ctx, incoming); err != nil {
				return err
			}
			if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
				TenantId: incoming.TenantId, UserId: incoming.UserId,
				ContractId: incoming.ContractId, OrderId: incoming.Id,
				EventType: "ORDER_CANCELED", Reason: controlReasonStaleMark,
				Detail:     "fresh positive mark price is required before matching",
				OperatorId: incoming.UserId, CreateTimes: now,
			}); err != nil {
				return err
			}
			changedOrders[incoming.Id] = incoming
			*order = *incoming
			return nil
		}
		if _, _, withinBand := optionOrderPriceBand(
			incoming.Price, lockedMarket.MarkPrice, lockedContract.OrderPriceBandRatio,
		); !withinBand {
			if err := cancelImmediateOrder(
				ctx, positionModel, instructionModel, incoming, controlReasonPriceBand, now,
			); err != nil {
				return err
			}
			if err := orderModel.Update(ctx, incoming); err != nil {
				return err
			}
			if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
				TenantId: incoming.TenantId, UserId: incoming.UserId,
				ContractId: incoming.ContractId, OrderId: incoming.Id,
				EventType: "ORDER_CANCELED", Reason: controlReasonPriceBand,
				Detail:     fmt.Sprintf("price=%s mark=%s", incoming.Price, lockedMarket.MarkPrice),
				OperatorId: incoming.UserId, CreateTimes: now,
			}); err != nil {
				return err
			}
			changedOrders[incoming.Id] = incoming
			*order = *incoming
			return nil
		}
		if incomingControl.KillSwitch == int64(common.YesNo_YES_NO_YES) {
			if err := cancelImmediateOrder(
				ctx, positionModel, instructionModel, incoming, controlReasonKillSwitch, now,
			); err != nil {
				return err
			}
			if err := orderModel.Update(ctx, incoming); err != nil {
				return err
			}
			if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
				TenantId: incoming.TenantId, UserId: incoming.UserId,
				ContractId: incoming.ContractId, OrderId: incoming.Id,
				EventType: "ORDER_CANCELED", Reason: controlReasonKillSwitch,
				Detail:     fmt.Sprintf("activated_at=%d", incomingControl.ActivatedAt),
				OperatorId: incoming.UserId, CreateTimes: now,
			}); err != nil {
				return err
			}
			changedOrders[incoming.Id] = incoming
			*order = *incoming
			return nil
		}
		if lockedContract.TenantId != incoming.TenantId ||
			lockedContract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) ||
			lockedContract.IsDeleted == int64(common.YesNo_YES_NO_YES) ||
			now < lockedContract.ListTime ||
			(lockedContract.ExpireTime > 0 && now >= lockedContract.ExpireTime) {
			if err := cancelImmediateOrder(
				ctx, positionModel, instructionModel, incoming, controlReasonContractClosed, now,
			); err != nil {
				return err
			}
			if err := orderModel.Update(ctx, incoming); err != nil {
				return err
			}
			if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
				TenantId: incoming.TenantId, UserId: incoming.UserId,
				ContractId: incoming.ContractId, OrderId: incoming.Id,
				EventType: "ORDER_CANCELED", Reason: controlReasonContractClosed,
				Detail:     fmt.Sprintf("contract_status=%d", lockedContract.Status),
				OperatorId: incoming.UserId, CreateTimes: now,
			}); err != nil {
				return err
			}
			changedOrders[incoming.Id] = incoming
			*order = *incoming
			return nil
		}
		contract = lockedContract
		mmpConfigModel := models.NewTOptionMmpConfigModel(conn, l.svcCtx.Config.CacheRedis)
		if incoming.Mmp == int64(common.YesNo_YES_NO_YES) {
			config, configErr := mmpConfigModel.FindForUpdate(
				ctx, incoming.TenantId, incoming.UserId, incoming.ContractId, incoming.MmpGroup,
			)
			if configErr != nil && !errors.Is(configErr, models.ErrNotFound) {
				return configErr
			}
			if configErr != nil ||
				config.Enabled != int64(common.YesNo_YES_NO_YES) ||
				config.Status != int64(option.MMPStatus_MMP_STATUS_ACTIVE) {
				reason := controlReasonMMPNotConfigured
				if configErr == nil {
					reason = controlReasonMMPDisabled
					if config.Status == int64(option.MMPStatus_MMP_STATUS_TRIGGERED) {
						reason = controlReasonMMPTriggered
					}
				}
				if err := cancelImmediateOrder(
					ctx, positionModel, instructionModel, incoming, reason, now,
				); err != nil {
					return err
				}
				if err := orderModel.Update(ctx, incoming); err != nil {
					return err
				}
				if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
					TenantId: incoming.TenantId, UserId: incoming.UserId,
					ContractId: incoming.ContractId, OrderId: incoming.Id,
					EventType: controlEventMMPOrderCanceled, Reason: reason,
					Detail:     fmt.Sprintf("group=%s", incoming.MmpGroup),
					OperatorId: incoming.UserId, CreateTimes: now,
				}); err != nil {
					return err
				}
				changedOrders[incoming.Id] = incoming
				*order = *incoming
				return nil
			}
		}
		selfOrders, err := orderModel.FindCrossingSelfOrders(
			ctx, incoming.TenantId, incoming.UserId, incoming.ContractId,
			oppositeOrderSide(incoming.Side), incoming.Price,
		)
		if err != nil {
			return err
		}
		if len(selfOrders) > 0 {
			now := time.Now().Unix()
			if err := cancelImmediateOrder(
				ctx, positionModel, instructionModel, incoming, controlReasonSelfTrade, now,
			); err != nil {
				return err
			}
			if err := orderModel.Update(ctx, incoming); err != nil {
				return err
			}
			if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
				TenantId: incoming.TenantId, UserId: incoming.UserId,
				ContractId: incoming.ContractId, OrderId: incoming.Id,
				EventType: controlEventSTPPrevented, Reason: controlReasonSelfTrade,
				Detail:     fmt.Sprintf("crossing_self_order_id=%d count=%d", selfOrders[0].Id, len(selfOrders)),
				OperatorId: incoming.UserId, CreateTimes: now,
			}); err != nil {
				return err
			}
			changedOrders[incoming.Id] = incoming
			*order = *incoming
			return nil
		}
		mmpConfigs := make(map[mmpGroupKey]*models.TOptionMmpConfig)
		if incoming.OrderType == int64(option.OrderType_ORDER_TYPE_POST_ONLY) ||
			incoming.OrderType == int64(option.OrderType_ORDER_TYPE_FOK) {
			candidates, err := orderModel.FindAllMatchableOrders(ctx, incoming.TenantId, incoming.ContractId, oppositeOrderSide(incoming.Side), incoming.UserId, incoming.AccountId, incoming.Price)
			if err != nil {
				return err
			}
			validCandidates := make([]*models.TOptionOrder, 0, len(candidates))
			for _, candidate := range candidates {
				if candidate.Mmp != int64(common.YesNo_YES_NO_YES) {
					validCandidates = append(validCandidates, candidate)
					continue
				}
				key := mmpGroupKey{
					tenantID: candidate.TenantId, userID: candidate.UserId,
					contractID: candidate.ContractId, groupCode: candidate.MmpGroup,
				}
				config := mmpConfigs[key]
				if config == nil {
					config, err = mmpConfigModel.FindForUpdate(
						ctx, candidate.TenantId, candidate.UserId,
						candidate.ContractId, candidate.MmpGroup,
					)
					if err != nil && !errors.Is(err, models.ErrNotFound) {
						return err
					}
					if err == nil {
						mmpConfigs[key] = config
					}
				}
				if config != nil &&
					config.Enabled == int64(common.YesNo_YES_NO_YES) &&
					config.Status == int64(option.MMPStatus_MMP_STATUS_ACTIVE) {
					validCandidates = append(validCandidates, candidate)
					continue
				}
				reason := controlReasonMMPNotConfigured
				if config != nil {
					reason = controlReasonMMPDisabled
					if config.Status == int64(option.MMPStatus_MMP_STATUS_TRIGGERED) {
						reason = controlReasonMMPTriggered
					}
				}
				if err := cancelImmediateOrder(
					ctx, positionModel, instructionModel, candidate, reason, now,
				); err != nil {
					return err
				}
				if err := orderModel.Update(ctx, candidate); err != nil {
					return err
				}
				candidateCopy := *candidate
				changedOrders[candidate.Id] = &candidateCopy
				if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
					TenantId: candidate.TenantId, UserId: candidate.UserId,
					ContractId: candidate.ContractId, OrderId: candidate.Id,
					EventType: controlEventMMPOrderCanceled, Reason: reason,
					Detail:     fmt.Sprintf("group=%s pre_match_filter=true", candidate.MmpGroup),
					OperatorId: candidate.UserId, CreateTimes: now,
				}); err != nil {
					return err
				}
			}
			matchableQty := matchableOptionQty(incoming, validCandidates)
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
		makerKillSwitch := make(map[int64]bool)
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
				if maker.UserId == incoming.UserId {
					continue
				}
				if _, _, withinBand := optionOrderPriceBand(
					maker.Price, lockedMarket.MarkPrice, lockedContract.OrderPriceBandRatio,
				); !withinBand {
					cancelTime := time.Now().Unix()
					if err := cancelImmediateOrder(
						ctx, positionModel, instructionModel, maker, controlReasonPriceBand, cancelTime,
					); err != nil {
						return err
					}
					if err := orderModel.Update(ctx, maker); err != nil {
						return err
					}
					if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
						TenantId: maker.TenantId, UserId: maker.UserId,
						ContractId: maker.ContractId, OrderId: maker.Id,
						EventType: "ORDER_CANCELED", Reason: controlReasonPriceBand,
						Detail:     fmt.Sprintf("maker_price=%s mark=%s", maker.Price, lockedMarket.MarkPrice),
						OperatorId: maker.UserId, CreateTimes: cancelTime,
					}); err != nil {
						return err
					}
					makerCopy := *maker
					changedOrders[maker.Id] = &makerCopy
					matched = true
					continue
				}
				killed, checked := makerKillSwitch[maker.UserId]
				if !checked {
					makerControl, controlErr := userControlModel.EnsureForUpdate(
						ctx, maker.TenantId, maker.UserId, time.Now().Unix(),
					)
					if controlErr != nil {
						return controlErr
					}
					killed = makerControl.KillSwitch == int64(common.YesNo_YES_NO_YES)
					makerKillSwitch[maker.UserId] = killed
				}
				if killed {
					cancelTime := time.Now().Unix()
					if err := cancelImmediateOrder(
						ctx, positionModel, instructionModel, maker, controlReasonKillSwitch, cancelTime,
					); err != nil {
						return err
					}
					if err := orderModel.Update(ctx, maker); err != nil {
						return err
					}
					if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
						TenantId: maker.TenantId, UserId: maker.UserId,
						ContractId: maker.ContractId, OrderId: maker.Id,
						EventType: "ORDER_CANCELED", Reason: controlReasonKillSwitch,
						Detail:     "maker excluded before matching",
						OperatorId: maker.UserId, CreateTimes: cancelTime,
					}); err != nil {
						return err
					}
					makerCopy := *maker
					changedOrders[maker.Id] = &makerCopy
					matched = true
					continue
				}
				var makerMMPConfig *models.TOptionMmpConfig
				if maker.Mmp == int64(common.YesNo_YES_NO_YES) {
					key := mmpGroupKey{
						tenantID: maker.TenantId, userID: maker.UserId,
						contractID: maker.ContractId, groupCode: maker.MmpGroup,
					}
					makerMMPConfig = mmpConfigs[key]
					if makerMMPConfig == nil {
						config, configErr := mmpConfigModel.FindForUpdate(
							ctx, maker.TenantId, maker.UserId, maker.ContractId, maker.MmpGroup,
						)
						if configErr != nil && !errors.Is(configErr, models.ErrNotFound) {
							return configErr
						}
						if configErr == nil {
							makerMMPConfig = config
							mmpConfigs[key] = config
						}
					}
					if makerMMPConfig == nil ||
						makerMMPConfig.Enabled != int64(common.YesNo_YES_NO_YES) ||
						makerMMPConfig.Status != int64(option.MMPStatus_MMP_STATUS_ACTIVE) {
						reason := controlReasonMMPNotConfigured
						if makerMMPConfig != nil {
							reason = controlReasonMMPDisabled
							if makerMMPConfig.Status == int64(option.MMPStatus_MMP_STATUS_TRIGGERED) {
								reason = controlReasonMMPTriggered
							}
						}
						cancelTime := time.Now().Unix()
						if err := cancelImmediateOrder(
							ctx, positionModel, instructionModel, maker, reason, cancelTime,
						); err != nil {
							return err
						}
						if err := orderModel.Update(ctx, maker); err != nil {
							return err
						}
						if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
							TenantId: maker.TenantId, UserId: maker.UserId,
							ContractId: maker.ContractId, OrderId: maker.Id,
							EventType: controlEventMMPOrderCanceled, Reason: reason,
							Detail:     fmt.Sprintf("group=%s", maker.MmpGroup),
							OperatorId: maker.UserId, CreateTimes: cancelTime,
						}); err != nil {
							return err
						}
						makerCopy := *maker
						changedOrders[maker.Id] = &makerCopy
						matched = true
						continue
					}
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
						FreezeBizNo: sellOrder.OrderNo, CollateralCoin: OptionOrderMarginCoin(sellOrder),
						Quantity: tradeQty, RemainingQuantity: tradeQty,
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
				if makerMMPConfig != nil {
					triggered, triggerReason := applyMMPFill(
						makerMMPConfig, maker.Side, tradePrice, tradeQty,
						lockedMarket.MarkPrice, optionMultiplier(contract),
						MMPMakerFee(trade, maker.Side), now,
					)
					if err := mmpConfigModel.Update(ctx, makerMMPConfig); err != nil {
						return err
					}
					if triggered {
						key := mmpGroupKey{
							tenantID: maker.TenantId, userID: maker.UserId,
							contractID: maker.ContractId, groupCode: maker.MmpGroup,
						}
						triggeredMMPGroups[key] = struct{}{}
						if _, err := controlEventModel.Insert(ctx, &models.TOptionTradingControlEvent{
							TenantId: maker.TenantId, UserId: maker.UserId,
							ContractId: maker.ContractId, OrderId: maker.Id,
							EventType: controlEventMMPTriggered, Reason: triggerReason,
							Detail: fmt.Sprintf(
								"group=%s qty=%s count=%d loss=%s mark=%s trade_price=%s",
								maker.MmpGroup, makerMMPConfig.AccumulatedQty,
								makerMMPConfig.TradeCount, makerMMPConfig.AccumulatedLoss,
								lockedMarket.MarkPrice, tradePrice,
							),
							OperatorId: maker.UserId, CreateTimes: now,
						}); err != nil {
							return err
						}
						if maker.UnfilledQty.IsPositive() {
							if err := cancelImmediateOrder(
								ctx, positionModel, instructionModel, maker, controlReasonMMPTriggered, now,
							); err != nil {
								return err
							}
						}
					}
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
				if err := updateMarketLastTrade(ctx, marketModel, lockedMarket, tradePrice, now); err != nil {
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
	var mmpCancelErr error
	for key := range triggeredMMPGroups {
		count, cancelErr := CancelMMPGroupOrders(
			l.ctx, l.svcCtx, key.tenantID, key.userID, key.contractID,
			key.groupCode, controlReasonMMPTriggered,
		)
		if cancelErr != nil {
			SetMMPConfigLastError(
				l.ctx, l.svcCtx, key.tenantID, key.userID, key.contractID,
				key.groupCode, cancelErr.Error(),
			)
			if mmpCancelErr == nil {
				mmpCancelErr = cancelErr
			}
			continue
		}
		if eventErr := insertTradingControlEvent(
			l.ctx, l.svcCtx, l.svcCtx.DB, &models.TOptionTradingControlEvent{
				TenantId: key.tenantID, UserId: key.userID, ContractId: key.contractID,
				EventType: controlEventMMPOrderCanceled, Reason: controlReasonMMPTriggered,
				Detail:     fmt.Sprintf("group=%s canceled=%d", key.groupCode, count),
				OperatorId: key.userID, CreateTimes: time.Now().Unix(),
			},
		); eventErr != nil && mmpCancelErr == nil {
			mmpCancelErr = eventErr
		}
	}
	for _, changedOrder := range changedOrders {
		publishOptionOrderChanged(l.ctx, l.svcCtx, changedOrder)
	}
	return mmpCancelErr
}

func rejectComboChildFromSimpleMatcher(order *models.TOptionOrder, path string) error {
	if order == nil || order.ComboOrderId == 0 {
		return nil
	}
	observability.RecordComboIsolationViolation(order.TenantId, path)
	return errors.New("combo child order cannot enter the simple order matcher")
}

func matchableOptionQty(incoming *models.TOptionOrder, candidates []*models.TOptionOrder) decimal.Decimal {
	total := decimal.Zero
	for _, candidate := range candidates {
		if candidate.Id == incoming.Id || !candidate.UnfilledQty.IsPositive() ||
			candidate.UserId == incoming.UserId {
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
			TargetBizNo: order.OrderNo, Coin: OptionOrderMarginCoin(order), Amount: order.MarginAmount,
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

func updateMarketLastTrade(
	ctx context.Context,
	model models.TOptionMarketModel,
	market *models.TOptionMarket,
	price decimal.Decimal,
	now int64,
) error {
	if market == nil {
		return errors.New("locked option market is missing")
	}

	market.LastPrice = price
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

func OptionOrderMarginCoin(order *models.TOptionOrder) string {
	if order == nil {
		return ""
	}
	// MarginCoin records the asset that was actually frozen. FeeCoin is not a
	// safe fallback: a covered physical Call freezes the underlying asset while
	// fees are denominated in the settlement coin.
	return strings.TrimSpace(order.MarginCoin)
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
			TargetBizNo: buyOrder.OrderNo, Coin: OptionOrderMarginCoin(buyOrder), Amount: buyOrder.MarginAmount,
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
