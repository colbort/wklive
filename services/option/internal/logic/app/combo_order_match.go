package applogic

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"wklive/common/generate"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var errComboLegLimitNotExecutable = errors.New("combo maker leg price violates taker leg limit")

// MatchFundedComboOrder matches one strategy order against the best
// price-time-priority inverse strategy. Every leg trade, order update, outbox
// event, margin lot and asset instruction is committed in one DB transaction.
func MatchFundedComboOrder(
	ctx context.Context, svcCtx *svc.ServiceContext, parent *models.TOptionComboOrder,
) error {
	if parent == nil ||
		(parent.Status != int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE) &&
			parent.Status != int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_PART_FILLED)) ||
		!parent.UnfilledQty.IsPositive() {
		return nil
	}
	candidates, err := svcCtx.OptionComboOrderModel.FindMatchCandidates(
		ctx, parent.TenantId, parent.InverseStrategyKey, 0,
		parent.NetPrice.Neg(), 50,
	)
	if err != nil {
		return err
	}
	var candidate *models.TOptionComboOrder
	selfCross := false
	for _, item := range candidates {
		if item.Id == parent.Id {
			continue
		}
		if item.UserId == parent.UserId {
			selfCross = true
			break
		}
		candidate = item
		break
	}
	if selfCross {
		return cancelRestingCombo(ctx, svcCtx, parent.Id, "SELF_TRADE_PREVENTED")
	}
	if candidate != nil &&
		parent.OrderType == int64(option.ComboOrderType_COMBO_ORDER_TYPE_FOK) &&
		candidate.UnfilledQty.LessThan(parent.UnfilledQty) {
		candidate = nil
	}
	if candidate == nil {
		if parent.OrderType == int64(option.ComboOrderType_COMBO_ORDER_TYPE_FOK) {
			return cancelRestingCombo(ctx, svcCtx, parent.Id, "FOK_NOT_FILLED")
		}
		return nil
	}
	matchNo, err := generate.GenerateNo(
		svcCtx.Redis, ctx, "combo_match_id", "OCM", "",
	)
	if err != nil {
		return err
	}
	changedChildren := make(map[int64]*models.TOptionOrder)
	err = svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		comboModel := models.NewTOptionComboOrderModel(conn, svcCtx.Config.CacheRedis)
		comboLegModel := models.NewTOptionComboOrderLegModel(conn, svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, svcCtx.Config.CacheRedis)
		marketModel := models.NewTOptionMarketModel(conn, svcCtx.Config.CacheRedis)
		haltModel := models.NewTOptionTradingHaltModel(conn, svcCtx.Config.CacheRedis)
		calendarModel := models.NewTOptionTradingCalendarModel(conn, svcCtx.Config.CacheRedis)
		calendarSessionModel := models.NewTOptionTradingCalendarSessionModel(conn, svcCtx.Config.CacheRedis)
		calendarExceptionModel := models.NewTOptionTradingCalendarExceptionModel(conn, svcCtx.Config.CacheRedis)
		userControlModel := models.NewTOptionUserTradingControlModel(conn, svcCtx.Config.CacheRedis)
		tradeModel := models.NewTOptionTradeModel(conn, svcCtx.Config.CacheRedis)
		matchSequenceModel := models.NewTOptionMatchSequenceModel(conn, svcCtx.Config.CacheRedis)
		outboxModel := models.NewTOptionOutboxModel(conn, svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, svcCtx.Config.CacheRedis)

		parentIDs := []int64{parent.Id, candidate.Id}
		sort.Slice(parentIDs, func(i, j int) bool { return parentIDs[i] < parentIDs[j] })
		lockedParents := make(map[int64]*models.TOptionComboOrder, 2)
		for _, id := range parentIDs {
			item, lockErr := comboModel.FindOneForUpdate(txCtx, id)
			if lockErr != nil {
				return lockErr
			}
			lockedParents[id] = item
		}
		incoming := lockedParents[parent.Id]
		maker := lockedParents[candidate.Id]
		if !comboParentMatchable(incoming) {
			return nil
		}
		if !comboParentMatchable(maker) ||
			maker.StrategyKey != incoming.InverseStrategyKey ||
			maker.UserId == incoming.UserId ||
			maker.NetPrice.LessThan(incoming.NetPrice.Neg()) {
			if incoming.OrderType == int64(option.ComboOrderType_COMBO_ORDER_TYPE_FOK) {
				children, childErr := orderModel.FindComboChildrenForUpdate(
					txCtx, incoming.TenantId, incoming.Id,
				)
				if childErr != nil {
					return childErr
				}
				return cancelComboInsideTx(
					txCtx, comboModel, orderModel, instructionModel,
					incoming, children, "FOK_NOT_FILLED", time.Now().Unix(),
				)
			}
			return nil
		}
		if incoming.OrderType == int64(option.ComboOrderType_COMBO_ORDER_TYPE_FOK) &&
			maker.UnfilledQty.LessThan(incoming.UnfilledQty) {
			children, childErr := orderModel.FindComboChildrenForUpdate(
				txCtx, incoming.TenantId, incoming.Id,
			)
			if childErr != nil {
				return childErr
			}
			return cancelComboInsideTx(
				txCtx, comboModel, orderModel, instructionModel,
				incoming, children, "FOK_NOT_FILLED", time.Now().Unix(),
			)
		}

		childrenByParent := make(map[int64][]*models.TOptionOrder, 2)
		legsByParent := make(map[int64][]*models.TOptionComboOrderLeg, 2)
		for _, id := range parentIDs {
			item := lockedParents[id]
			children, childErr := orderModel.FindComboChildrenForUpdate(
				txCtx, item.TenantId, item.Id,
			)
			if childErr != nil {
				return childErr
			}
			legs, legErr := comboLegModel.FindByComboOrderIDForUpdate(
				txCtx, item.TenantId, item.Id,
			)
			if legErr != nil {
				return legErr
			}
			if len(children) < minComboLegs || len(children) > maxComboLegs ||
				len(children) != len(legs) {
				return errors.New("combo leg/child cardinality invariant violated")
			}
			childrenByParent[id] = children
			legsByParent[id] = legs
		}

		userIDs := []int64{incoming.UserId, maker.UserId}
		sort.Slice(userIDs, func(i, j int) bool { return userIDs[i] < userIDs[j] })
		for _, userID := range userIDs {
			control, controlErr := userControlModel.EnsureForUpdate(
				txCtx, incoming.TenantId, userID, time.Now().Unix(),
			)
			if controlErr != nil {
				return controlErr
			}
			if control.KillSwitch == int64(common.YesNo_YES_NO_YES) {
				target := incoming
				if userID == maker.UserId {
					target = maker
				}
				return cancelComboInsideTx(
					txCtx, comboModel, orderModel, instructionModel,
					target, childrenByParent[target.Id],
					controlReasonKillSwitch, time.Now().Unix(),
				)
			}
		}

		incomingLegs := legsByParent[incoming.Id]
		makerLegs := legsByParent[maker.Id]
		incomingChildren := childrenByParent[incoming.Id]
		makerChildren := childrenByParent[maker.Id]
		contracts := make(map[int64]*models.TOptionContract, len(incomingLegs))
		markets := make(map[int64]*models.TOptionMarket, len(incomingLegs))
		now := time.Now().Unix()
		for index, incomingLeg := range incomingLegs {
			makerLeg := makerLegs[index]
			incomingChild := incomingChildren[index]
			makerChild := makerChildren[index]
			if err := validateInverseComboLegs(
				incoming, maker, incomingLeg, makerLeg, incomingChild, makerChild,
			); err != nil {
				if errors.Is(err, errComboLegLimitNotExecutable) {
					if incoming.OrderType == int64(option.ComboOrderType_COMBO_ORDER_TYPE_FOK) {
						return cancelComboInsideTx(
							txCtx, comboModel, orderModel, instructionModel,
							incoming, incomingChildren, "FOK_NOT_FILLED", now,
						)
					}
					return nil
				}
				return err
			}
			contract, lockErr := contractModel.FindOneForUpdate(txCtx, incomingLeg.ContractId)
			if lockErr != nil {
				return lockErr
			}
			market, marketErr := marketModel.FindOneByTenantIdContractIdForUpdate(
				txCtx, incoming.TenantId, incomingLeg.ContractId,
			)
			if marketErr != nil {
				return marketErr
			}
			if contract.TenantId != incoming.TenantId ||
				contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) ||
				contract.IsDeleted == int64(common.YesNo_YES_NO_YES) ||
				now < contract.ListTime ||
				(contract.ExpireTime > 0 && now >= contract.ExpireTime) ||
				!logichelpers.IsMarkFresh(market, now, 30) ||
				!market.MarkPrice.IsPositive() {
				return cancelComboInsideTx(
					txCtx, comboModel, orderModel, instructionModel,
					incoming, incomingChildren, controlReasonContractClosed, now,
				)
			}
			calendarDecision, calendarErr := logichelpers.IsContractTradingOpenWithModels(
				txCtx, haltModel, calendarModel, calendarSessionModel,
				calendarExceptionModel, contract, now,
			)
			if calendarErr != nil || calendarDecision == nil || !calendarDecision.Open {
				return cancelComboInsideTx(
					txCtx, comboModel, orderModel, instructionModel,
					incoming, incomingChildren, controlReasonContractClosed, now,
				)
			}
			if _, _, ok := optionOrderPriceBand(
				makerChild.Price, market.MarkPrice, contract.OrderPriceBandRatio,
			); !ok {
				return cancelComboInsideTx(
					txCtx, comboModel, orderModel, instructionModel,
					maker, makerChildren, controlReasonPriceBand, now,
				)
			}
			contracts[contract.Id] = contract
			markets[contract.Id] = market
		}

		comboQty := decimal.Min(incoming.UnfilledQty, maker.UnfilledQty)
		if !comboQty.IsPositive() {
			return nil
		}
		for index, incomingLeg := range incomingLegs {
			makerLeg := makerLegs[index]
			incomingChild := incomingChildren[index]
			makerChild := makerChildren[index]
			contract := contracts[incomingLeg.ContractId]
			market := markets[incomingLeg.ContractId]
			legQty := comboQty.Mul(decimal.NewFromInt(incomingLeg.Ratio))
			tradePrice := makerLeg.Price
			tradeNo := fmt.Sprintf("%s-L%d", matchNo, incomingLeg.LegNo)
			matchSequence, sequenceErr := matchSequenceModel.Next(
				txCtx, incoming.TenantId, incomingLeg.ContractId,
			)
			if sequenceErr != nil {
				return sequenceErr
			}
			buyOrder, sellOrder := incomingChild, makerChild
			if incomingChild.Side == int64(common.Side_SIDE_SELL) {
				buyOrder, sellOrder = makerChild, incomingChild
			}
			trade := &models.TOptionTrade{
				TenantId: incoming.TenantId, TradeNo: tradeNo,
				ComboMatchNo: matchNo, ComboLegNo: incomingLeg.LegNo,
				ContractId:       incomingLeg.ContractId,
				UnderlyingSymbol: contract.UnderlyingSymbol,
				BuyOrderId:       buyOrder.Id, BuyOrderNo: buyOrder.OrderNo,
				BuyUserId: buyOrder.UserId, BuyAccountId: buyOrder.AccountId,
				SellOrderId: sellOrder.Id, SellOrderNo: sellOrder.OrderNo,
				SellUserId: sellOrder.UserId, SellAccountId: sellOrder.AccountId,
				Price: tradePrice, Qty: legQty,
				Turnover: optionTurnover(contract, tradePrice, legQty),
				FeeCoin:  contract.SettleCoin, MakerSide: makerSide(makerChild.Side),
				MatchSequence: matchSequence, TradeTime: now, CreateTimes: now,
			}
			trade.BuyFee, trade.SellFee = optionTradeFees(
				contract, trade.Turnover, trade.MakerSide,
			)
			requiredBuyReservation := trade.Turnover.Add(trade.BuyFee)
			if buyOrder.MarginAmount.LessThan(requiredBuyReservation) {
				return errors.New("combo buy-leg reservation is insufficient for atomic execution")
			}
			tradeResult, insertErr := tradeModel.Insert(txCtx, trade)
			if insertErr != nil {
				return insertErr
			}
			trade.Id, insertErr = tradeResult.LastInsertId()
			if insertErr != nil {
				return insertErr
			}
			sellerMargin := allocateSellerMargin(
				sellOrder, legQty, sellOrder.UnfilledQty,
			)
			if sellerMargin.IsPositive() {
				if _, insertErr = marginLotModel.Insert(txCtx, &models.TOptionMarginLot{
					TenantId: sellOrder.TenantId, UserId: sellOrder.UserId,
					AccountId: sellOrder.AccountId, ContractId: sellOrder.ContractId,
					OrderId: sellOrder.Id, TradeId: trade.Id,
					FreezeBizNo:    sellOrder.OrderNo,
					CollateralCoin: OptionOrderMarginCoin(sellOrder),
					Quantity:       legQty, RemainingQuantity: legQty,
					InitialMargin: sellerMargin, RemainingMargin: sellerMargin,
					Status:      int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
					CreateTimes: now, UpdateTimes: now,
				}); insertErr != nil {
					return insertErr
				}
			}
			if _, insertErr = outboxModel.Insert(txCtx, &models.TOptionOutbox{
				TenantId: incoming.TenantId, EventNo: trade.TradeNo + "-POSITION",
				EventType:  int64(option.OptionEventType_OPTION_EVENT_TYPE_TRADE_POSITION),
				ContractId: trade.ContractId, TradeId: trade.Id,
				MatchSequence: trade.MatchSequence,
				Status:        int64(option.OptionEventStatus_OPTION_EVENT_STATUS_PENDING),
				CreateTimes:   now, UpdateTimes: now,
			}); insertErr != nil {
				return insertErr
			}

			applyTradeToOrder(incomingChild, contract, tradePrice, legQty, now)
			applyTradeToOrder(makerChild, contract, tradePrice, legQty, now)
			buyOrder.Fee = buyOrder.Fee.Add(trade.BuyFee)
			sellOrder.Fee = sellOrder.Fee.Add(trade.SellFee)
			consumeBuyOrderReservation(buyOrder, trade.Turnover.Add(trade.BuyFee))
			if insertErr = createTradeAssetInstructions(
				txCtx, instructionModel, contract, trade, buyOrder, sellOrder, now,
			); insertErr != nil {
				return insertErr
			}
			if insertErr = orderModel.Update(txCtx, incomingChild); insertErr != nil {
				return insertErr
			}
			if insertErr = orderModel.Update(txCtx, makerChild); insertErr != nil {
				return insertErr
			}
			applyComboLegFill(incomingLeg, legQty, now)
			applyComboLegFill(makerLeg, legQty, now)
			if insertErr = comboLegModel.Update(txCtx, incomingLeg); insertErr != nil {
				return insertErr
			}
			if insertErr = comboLegModel.Update(txCtx, makerLeg); insertErr != nil {
				return insertErr
			}
			if insertErr = updateMarketLastTrade(
				txCtx, marketModel, market, tradePrice, now,
			); insertErr != nil {
				return insertErr
			}
			incomingCopy := *incomingChild
			makerCopy := *makerChild
			changedChildren[incomingChild.Id] = &incomingCopy
			changedChildren[makerChild.Id] = &makerCopy
		}
		applyComboParentFill(incoming, comboQty, now)
		applyComboParentFill(maker, comboQty, now)
		if err := comboModel.Update(txCtx, incoming); err != nil {
			return err
		}
		if err := comboModel.Update(txCtx, maker); err != nil {
			return err
		}
		*parent = *incoming
		return nil
	})
	if err != nil {
		return err
	}
	for _, child := range changedChildren {
		publishOptionOrderChanged(ctx, svcCtx, child)
	}
	return nil
}

func comboParentMatchable(item *models.TOptionComboOrder) bool {
	if item == nil || !item.UnfilledQty.IsPositive() {
		return false
	}
	return item.Status == int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE) ||
		item.Status == int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_PART_FILLED)
}

func validateInverseComboLegs(
	incomingParent, makerParent *models.TOptionComboOrder,
	incomingLeg, makerLeg *models.TOptionComboOrderLeg,
	incomingChild, makerChild *models.TOptionOrder,
) error {
	if incomingLeg == nil || makerLeg == nil || incomingChild == nil || makerChild == nil {
		return errors.New("combo leg mapping is incomplete")
	}
	if incomingLeg.ContractId != makerLeg.ContractId ||
		incomingLeg.LegNo != makerLeg.LegNo ||
		incomingLeg.Ratio != makerLeg.Ratio ||
		incomingLeg.Side == makerLeg.Side ||
		incomingLeg.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_OPEN) ||
		makerLeg.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_OPEN) {
		return errors.New("combo inverse-strategy invariant violated")
	}
	if (incomingLeg.Side == int64(common.Side_SIDE_BUY) &&
		makerLeg.Price.GreaterThan(incomingLeg.Price)) ||
		(incomingLeg.Side == int64(common.Side_SIDE_SELL) &&
			makerLeg.Price.LessThan(incomingLeg.Price)) {
		return errComboLegLimitNotExecutable
	}
	if incomingChild.ComboOrderId != incomingParent.Id ||
		makerChild.ComboOrderId != makerParent.Id ||
		incomingChild.ComboLegNo != incomingLeg.LegNo ||
		makerChild.ComboLegNo != makerLeg.LegNo ||
		incomingChild.Id != incomingLeg.ChildOrderId ||
		makerChild.Id != makerLeg.ChildOrderId ||
		incomingChild.ContractId != incomingLeg.ContractId ||
		makerChild.ContractId != makerLeg.ContractId ||
		incomingChild.Side != incomingLeg.Side ||
		makerChild.Side != makerLeg.Side ||
		!incomingChild.Price.Equal(incomingLeg.Price) ||
		!makerChild.Price.Equal(makerLeg.Price) {
		return errors.New("combo shadow-child invariant violated")
	}
	if (incomingChild.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) &&
		incomingChild.Status != int64(option.OrderStatus_ORDER_STATUS_PART_FILLED)) ||
		(makerChild.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) &&
			makerChild.Status != int64(option.OrderStatus_ORDER_STATUS_PART_FILLED)) {
		return errors.New("combo shadow child is not matchable")
	}
	return nil
}

func applyComboLegFill(
	leg *models.TOptionComboOrderLeg, qty decimal.Decimal, now int64,
) {
	leg.FilledQty = leg.FilledQty.Add(qty)
	leg.UnfilledQty = decimal.Max(leg.Qty.Sub(leg.FilledQty), decimal.Zero)
	leg.UpdateTimes = now
}

func applyComboParentFill(
	parent *models.TOptionComboOrder, qty decimal.Decimal, now int64,
) {
	parent.FilledQty = parent.FilledQty.Add(qty)
	parent.UnfilledQty = decimal.Max(parent.Qty.Sub(parent.FilledQty), decimal.Zero)
	if parent.UnfilledQty.IsPositive() {
		parent.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_PART_FILLED)
	} else {
		parent.UnfilledQty = decimal.Zero
		parent.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_FILLED)
	}
	parent.UpdateTimes = now
}

func cancelRestingCombo(
	ctx context.Context, svcCtx *svc.ServiceContext, comboOrderID int64, reason string,
) error {
	var changedChildren []*models.TOptionOrder
	err := svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		comboModel := models.NewTOptionComboOrderModel(conn, svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, svcCtx.Config.CacheRedis)
		parent, err := comboModel.FindOneForUpdate(txCtx, comboOrderID)
		if err != nil {
			return err
		}
		if !comboParentMatchable(parent) {
			return nil
		}
		children, err := orderModel.FindComboChildrenForUpdate(
			txCtx, parent.TenantId, parent.Id,
		)
		if err != nil {
			return err
		}
		if err = cancelComboInsideTx(
			txCtx, comboModel, orderModel, instructionModel,
			parent, children, reason, time.Now().Unix(),
		); err != nil {
			return err
		}
		for _, child := range children {
			childCopy := *child
			changedChildren = append(changedChildren, &childCopy)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, child := range changedChildren {
		publishOptionOrderChanged(ctx, svcCtx, child)
	}
	return nil
}

// CancelComboOrderByControl is shared by kill-switch, circuit-breaker,
// lifecycle and administrative cancellation paths. It never cancels only one
// shadow child.
func CancelComboOrderByControl(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	comboOrderID int64,
	reason string,
) error {
	var changedChildren []*models.TOptionOrder
	err := svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		comboModel := models.NewTOptionComboOrderModel(conn, svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, svcCtx.Config.CacheRedis)
		parent, err := comboModel.FindOneForUpdate(txCtx, comboOrderID)
		if err != nil {
			return err
		}
		switch option.ComboOrderStatus(parent.Status) {
		case option.ComboOrderStatus_COMBO_ORDER_STATUS_FUNDING:
			children, childErr := orderModel.FindComboChildrenForUpdate(
				txCtx, parent.TenantId, parent.Id,
			)
			if childErr != nil {
				return childErr
			}
			if len(children) < minComboLegs || len(children) > maxComboLegs {
				return errors.New("combo child-order cardinality invariant violated")
			}
			now := time.Now().Unix()
			requiresRelease := false
			for _, child := range children {
				freeze, findErr := instructionModel.FindOneByTenantIdInstructionNo(
					txCtx, child.TenantId, child.OrderNo+"-FREEZE",
				)
				if findErr != nil {
					return findErr
				}
				freeze, findErr = instructionModel.FindOneForUpdate(txCtx, freeze.Id)
				if findErr != nil {
					return findErr
				}
				cancelBeforeFreeze := false
				if freeze.Status == int64(
					option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING,
				) {
					freeze.Status = int64(
						option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_CANCELED,
					)
					freeze.UpdateTimes = now
					if findErr = instructionModel.Update(txCtx, freeze); findErr != nil {
						return findErr
					}
					cancelBeforeFreeze = true
					child.MarginAmount = decimal.Zero
				}
				child.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
				if child.MarginAmount.IsPositive() && !cancelBeforeFreeze {
					requiresRelease = true
					child.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELING)
					if _, findErr = instructionModel.Insert(
						txCtx, &models.TOptionAssetInstruction{
							TenantId:      child.TenantId,
							InstructionNo: child.OrderNo + "-COMBO-CONTROL-RELEASE",
							BizNo:         child.OrderNo, OrderId: child.Id,
							UserId: child.UserId, AccountId: child.AccountId,
							Action: int64(
								option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN,
							),
							TargetBizNo: child.OrderNo,
							Coin:        OptionOrderMarginCoin(child), Amount: child.MarginAmount,
							StepNo: 2,
							Status: int64(
								option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING,
							),
							ReconciliationStatus: int64(
								option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING,
							),
							CreateTimes: now, UpdateTimes: now,
						},
					); findErr != nil {
						return findErr
					}
				}
				child.CancelReason = reason
				child.CancelTime = now
				child.UpdateTimes = now
				if findErr = orderModel.Update(txCtx, child); findErr != nil {
					return findErr
				}
				childCopy := *child
				changedChildren = append(changedChildren, &childCopy)
			}
			parent.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED)
			if requiresRelease {
				parent.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELING)
			}
			parent.CancelReason = reason
			parent.CancelTime = now
			parent.UpdateTimes = now
			return comboModel.Update(txCtx, parent)
		case option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE,
			option.ComboOrderStatus_COMBO_ORDER_STATUS_PART_FILLED:
			children, childErr := orderModel.FindComboChildrenForUpdate(
				txCtx, parent.TenantId, parent.Id,
			)
			if childErr != nil {
				return childErr
			}
			if childErr = cancelComboInsideTx(
				txCtx, comboModel, orderModel, instructionModel,
				parent, children, reason, time.Now().Unix(),
			); childErr != nil {
				return childErr
			}
			for _, child := range children {
				childCopy := *child
				changedChildren = append(changedChildren, &childCopy)
			}
			return nil
		default:
			return nil
		}
	})
	if err != nil {
		return err
	}
	for _, child := range changedChildren {
		publishOptionOrderChanged(ctx, svcCtx, child)
	}
	return nil
}

func cancelComboInsideTx(
	ctx context.Context,
	comboModel models.TOptionComboOrderModel,
	orderModel models.TOptionOrderModel,
	instructionModel models.TOptionAssetInstructionModel,
	parent *models.TOptionComboOrder,
	children []*models.TOptionOrder,
	reason string,
	now int64,
) error {
	if len(children) < minComboLegs || len(children) > maxComboLegs {
		return errors.New("combo child-order cardinality invariant violated")
	}
	requiresRelease := false
	for _, child := range children {
		if child.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) &&
			child.Status != int64(option.OrderStatus_ORDER_STATUS_PART_FILLED) {
			return errors.New("combo child cannot be canceled from current state")
		}
		child.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
		if child.MarginAmount.IsPositive() {
			requiresRelease = true
			child.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELING)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId:      child.TenantId,
				InstructionNo: child.OrderNo + "-COMBO-MATCH-RELEASE",
				BizNo:         child.OrderNo, OrderId: child.Id, UserId: child.UserId,
				AccountId: child.AccountId,
				Action: int64(
					option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN,
				),
				TargetBizNo: child.OrderNo, Coin: OptionOrderMarginCoin(child),
				Amount: child.MarginAmount, StepNo: 3,
				Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(
					option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING,
				),
				CreateTimes: now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		child.CancelReason = reason
		child.CancelTime = now
		child.UpdateTimes = now
		if err := orderModel.Update(ctx, child); err != nil {
			return err
		}
	}
	parent.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED)
	if requiresRelease {
		parent.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELING)
	}
	parent.CancelReason = reason
	parent.CancelTime = now
	parent.UpdateTimes = now
	return comboModel.Update(ctx, parent)
}
