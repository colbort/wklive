package internallogic

import (
	"context"
	"errors"

	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func optionMultiplier(contract *models.TOptionContract) decimal.Decimal {
	if contract.Multiplier.IsPositive() {
		return contract.Multiplier
	}
	if contract.ContractUnit.IsPositive() {
		return contract.ContractUnit
	}
	return decimal.NewFromInt(1)
}

func optionTurnover(contract *models.TOptionContract, price, qty decimal.Decimal) decimal.Decimal {
	return price.Mul(qty).Mul(optionMultiplier(contract))
}

func oppositeOrderSide(side int64) int64 {
	if side == int64(common.Side_SIDE_BUY) {
		return int64(common.Side_SIDE_SELL)
	}
	if side == int64(common.Side_SIDE_SELL) {
		return int64(common.Side_SIDE_BUY)
	}
	return 0
}

func openPositionSide(orderSide int64) int64 {
	if orderSide == int64(common.Side_SIDE_BUY) {
		return int64(common.PositionSide_POSITION_SIDE_LONG)
	}
	if orderSide == int64(common.Side_SIDE_SELL) {
		return int64(common.PositionSide_POSITION_SIDE_SHORT)
	}
	return 0
}

func closePositionSide(orderSide int64) int64 {
	if orderSide == int64(common.Side_SIDE_SELL) {
		return int64(common.PositionSide_POSITION_SIDE_LONG)
	}
	if orderSide == int64(common.Side_SIDE_BUY) {
		return int64(common.PositionSide_POSITION_SIDE_SHORT)
	}
	return 0
}

func applyTradeToOrder(order *models.TOptionOrder, contract *models.TOptionContract, price, qty decimal.Decimal, now int64) {
	prevFilled := order.FilledQty
	nextFilled := prevFilled.Add(qty)
	if !nextFilled.IsPositive() {
		return
	}

	order.AvgPrice = order.AvgPrice.Mul(prevFilled).Add(price.Mul(qty)).Div(nextFilled)
	order.FilledQty = nextFilled
	order.UnfilledQty = decimal.Max(order.Qty.Sub(order.FilledQty), decimal.Zero)
	order.Turnover = order.Turnover.Add(optionTurnover(contract, price, qty))
	order.MatchTime = now
	order.UpdateTimes = now
	if !order.UnfilledQty.IsPositive() {
		order.UnfilledQty = decimal.Zero
		order.Status = int64(option.OrderStatus_ORDER_STATUS_FILLED)
		return
	}
	order.Status = int64(option.OrderStatus_ORDER_STATUS_PART_FILLED)
}

func updateOpenPosition(ctx context.Context, model models.TOptionPositionModel, contract *models.TOptionContract, order *models.TOptionOrder, price, qty decimal.Decimal, now int64) error {
	side := openPositionSide(order.Side)
	if side == 0 {
		return i18n.StatusError(ctx, i18n.ParamError)
	}

	pos, err := model.FindOneByTenantIdUserIdAccountIdContractIdSide(ctx, order.TenantId, order.UserId, order.AccountId, order.ContractId, side)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}

	multiplier := optionMultiplier(contract)
	if errors.Is(err, models.ErrNotFound) {
		exerciseableQty := decimal.Zero
		if side == int64(common.PositionSide_POSITION_SIDE_LONG) {
			exerciseableQty = qty
		}
		_, err = model.Insert(ctx, &models.TOptionPosition{
			TenantId:         order.TenantId,
			UserId:           order.UserId,
			AccountId:        order.AccountId,
			ContractId:       order.ContractId,
			UnderlyingSymbol: order.UnderlyingSymbol,
			Side:             side,
			PositionQty:      qty,
			AvailableQty:     qty,
			OpenAvgPrice:     price,
			MarkPrice:        price,
			PositionValue:    price.Mul(qty).Mul(multiplier),
			ExerciseableQty:  exerciseableQty,
			Status:           int64(option.PositionStatus_POSITION_STATUS_HOLDING),
			LastCalcTime:     now,
			CreateTimes:      now,
			UpdateTimes:      now,
		})
		return err
	}

	nextQty := pos.PositionQty.Add(qty)
	if !nextQty.IsPositive() {
		return nil
	}
	pos.OpenAvgPrice = pos.OpenAvgPrice.Mul(pos.PositionQty).Add(price.Mul(qty)).Div(nextQty)
	pos.PositionQty = nextQty
	pos.AvailableQty = pos.AvailableQty.Add(qty)
	pos.MarkPrice = price
	pos.PositionValue = pos.MarkPrice.Mul(pos.PositionQty).Mul(multiplier)
	if side == int64(common.PositionSide_POSITION_SIDE_LONG) {
		pos.ExerciseableQty = pos.ExerciseableQty.Add(qty)
		pos.UnrealizedPnl = pos.MarkPrice.Sub(pos.OpenAvgPrice).Mul(pos.PositionQty).Mul(multiplier)
	} else {
		pos.UnrealizedPnl = pos.OpenAvgPrice.Sub(pos.MarkPrice).Mul(pos.PositionQty).Mul(multiplier)
	}
	pos.Status = int64(option.PositionStatus_POSITION_STATUS_HOLDING)
	pos.LastCalcTime = now
	pos.UpdateTimes = now
	return model.Update(ctx, pos)
}

func updateClosePosition(ctx context.Context, model models.TOptionPositionModel, contract *models.TOptionContract, order *models.TOptionOrder, price, qty decimal.Decimal, now int64) error {
	side := closePositionSide(order.Side)
	if side == 0 {
		return i18n.StatusError(ctx, i18n.ParamError)
	}

	pos, err := model.FindOneByTenantIdUserIdAccountIdContractIdSide(ctx, order.TenantId, order.UserId, order.AccountId, order.ContractId, side)
	if err != nil {
		return err
	}
	if pos.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) {
		return i18n.StatusError(ctx, i18n.OperationNotAllowed)
	}

	reduceQty := decimal.Min(qty, pos.PositionQty)
	multiplier := optionMultiplier(contract)
	if side == int64(common.PositionSide_POSITION_SIDE_LONG) {
		pos.RealizedPnl = pos.RealizedPnl.Add(price.Sub(pos.OpenAvgPrice).Mul(reduceQty).Mul(multiplier))
		pos.ExerciseableQty = decimal.Max(pos.ExerciseableQty.Sub(reduceQty), decimal.Zero)
	} else {
		pos.RealizedPnl = pos.RealizedPnl.Add(pos.OpenAvgPrice.Sub(price).Mul(reduceQty).Mul(multiplier))
	}

	if pos.FrozenQty.GreaterThanOrEqual(reduceQty) {
		pos.FrozenQty = pos.FrozenQty.Sub(reduceQty)
	} else {
		left := reduceQty.Sub(pos.FrozenQty)
		pos.FrozenQty = decimal.Zero
		pos.AvailableQty = decimal.Max(pos.AvailableQty.Sub(left), decimal.Zero)
	}
	pos.PositionQty = decimal.Max(pos.PositionQty.Sub(reduceQty), decimal.Zero)
	pos.MarkPrice = price
	pos.PositionValue = pos.MarkPrice.Mul(pos.PositionQty).Mul(multiplier)
	pos.LastCalcTime = now
	pos.UpdateTimes = now
	if !pos.PositionQty.IsPositive() {
		pos.PositionQty = decimal.Zero
		pos.AvailableQty = decimal.Zero
		pos.FrozenQty = decimal.Zero
		pos.ExerciseableQty = decimal.Zero
		pos.UnrealizedPnl = decimal.Zero
		pos.PositionValue = decimal.Zero
		pos.Status = int64(option.PositionStatus_POSITION_STATUS_CLOSED)
	}
	return model.Update(ctx, pos)
}

func updatePositionByFilledOrder(ctx context.Context, model models.TOptionPositionModel, contract *models.TOptionContract, order *models.TOptionOrder, price, qty decimal.Decimal, now int64) error {
	if order.PositionEffect == int64(option.PositionEffect_POSITION_EFFECT_CLOSE) {
		return updateClosePosition(ctx, model, contract, order, price, qty, now)
	}
	return updateOpenPosition(ctx, model, contract, order, price, qty, now)
}

func freezeClosePosition(ctx context.Context, model models.TOptionPositionModel, order *models.TOptionOrder, now int64) error {
	if order.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_CLOSE) {
		return nil
	}

	side := closePositionSide(order.Side)
	if side == 0 {
		return i18n.StatusError(ctx, i18n.ParamError)
	}
	pos, err := model.FindOneByTenantIdUserIdAccountIdContractIdSide(ctx, order.TenantId, order.UserId, order.AccountId, order.ContractId, side)
	if err != nil {
		return err
	}
	if pos.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) || pos.AvailableQty.LessThan(order.Qty) {
		return i18n.StatusError(ctx, i18n.QuantityFormatError)
	}

	pos.AvailableQty = pos.AvailableQty.Sub(order.Qty)
	pos.FrozenQty = pos.FrozenQty.Add(order.Qty)
	pos.UpdateTimes = now
	return model.Update(ctx, pos)
}

func releaseClosePositionFrozenQty(ctx context.Context, model models.TOptionPositionModel, order *models.TOptionOrder, qty decimal.Decimal, now int64) error {
	if order.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_CLOSE) || !qty.IsPositive() {
		return nil
	}

	side := closePositionSide(order.Side)
	if side == 0 {
		return i18n.StatusError(ctx, i18n.ParamError)
	}
	pos, err := model.FindOneByTenantIdUserIdAccountIdContractIdSide(ctx, order.TenantId, order.UserId, order.AccountId, order.ContractId, side)
	if err != nil {
		return err
	}
	releaseQty := decimal.Min(qty, pos.FrozenQty)
	pos.FrozenQty = decimal.Max(pos.FrozenQty.Sub(releaseQty), decimal.Zero)
	pos.AvailableQty = pos.AvailableQty.Add(releaseQty)
	pos.UpdateTimes = now
	return model.Update(ctx, pos)
}

func optionIntrinsicValue(contract *models.TOptionContract, deliveryPrice decimal.Decimal) decimal.Decimal {
	if contract.OptionType == int64(option.OptionType_OPTION_TYPE_CALL) {
		return decimal.Max(deliveryPrice.Sub(contract.StrikePrice), decimal.Zero)
	}
	if contract.OptionType == int64(option.OptionType_OPTION_TYPE_PUT) {
		return decimal.Max(contract.StrikePrice.Sub(deliveryPrice), decimal.Zero)
	}
	return decimal.Zero
}

func optionSettlementPayoff(contract *models.TOptionContract, deliveryPrice, qty decimal.Decimal) decimal.Decimal {
	return optionIntrinsicValue(contract, deliveryPrice).Mul(qty).Mul(optionMultiplier(contract))
}

func optionExerciseAmount(contract *models.TOptionContract, qty decimal.Decimal) decimal.Decimal {
	return contract.StrikePrice.Mul(qty).Mul(optionMultiplier(contract))
}

func applyOptionAccountDelta(ctx context.Context, accountModel models.TOptionAccountModel, billModel models.TOptionBillModel, tenantId, userId, accountId int64, coin string, amount decimal.Decimal, refType, refId int64, bizNo, remark string, realized bool, now int64) error {
	if amount.IsZero() {
		return nil
	}

	account, err := accountModel.FindOneByTenantIdUserIdAccountIdMarginCoin(ctx, tenantId, userId, accountId, coin)
	before := decimal.Zero
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if errors.Is(err, models.ErrNotFound) {
		account = &models.TOptionAccount{
			TenantId:         tenantId,
			UserId:           userId,
			AccountId:        accountId,
			MarginCoin:       coin,
			Balance:          amount,
			AvailableBalance: amount,
			Status:           int64(option.AccountStatus_ACCOUNT_STATUS_NORMAL),
			CreateTimes:      now,
			UpdateTimes:      now,
		}
		if realized {
			account.RealizedPnl = amount
		}
		if _, err := accountModel.Insert(ctx, account); err != nil {
			return err
		}
	} else {
		before = account.Balance
		account.Balance = account.Balance.Add(amount)
		account.AvailableBalance = account.AvailableBalance.Add(amount)
		if realized {
			account.RealizedPnl = account.RealizedPnl.Add(amount)
		}
		account.UpdateTimes = now
		if err := accountModel.Update(ctx, account); err != nil {
			return err
		}
	}

	_, err = billModel.Insert(ctx, &models.TOptionBill{
		TenantId:      tenantId,
		UserId:        userId,
		AccountId:     accountId,
		BizNo:         bizNo,
		RefType:       refType,
		RefId:         refId,
		Coin:          coin,
		ChangeAmount:  amount,
		BalanceBefore: before,
		BalanceAfter:  before.Add(amount),
		Remark:        remark,
		CreateTimes:   now,
	})
	return err
}
