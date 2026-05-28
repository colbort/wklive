package logic

import (
	"context"
	"errors"
	"fmt"
	"math"

	"wklive/proto/option"
	"wklive/services/option/models"
)

const optionFloatEpsilon = 1e-9

var errInsufficientPositionQuantity = errors.New("insufficient position quantity")

func optionMultiplier(contract *models.TOptionContract) float64 {
	if contract.Multiplier > 0 {
		return contract.Multiplier
	}
	if contract.ContractUnit > 0 {
		return contract.ContractUnit
	}
	return 1
}

func optionTurnover(contract *models.TOptionContract, price, qty float64) float64 {
	return price * qty * optionMultiplier(contract)
}

func oppositeOrderSide(side int64) int64 {
	if side == int64(option.Side_SIDE_BUY) {
		return int64(option.Side_SIDE_SELL)
	}
	if side == int64(option.Side_SIDE_SELL) {
		return int64(option.Side_SIDE_BUY)
	}
	return 0
}

func openPositionSide(orderSide int64) int64 {
	if orderSide == int64(option.Side_SIDE_BUY) {
		return int64(option.PositionSide_POSITION_SIDE_LONG)
	}
	if orderSide == int64(option.Side_SIDE_SELL) {
		return int64(option.PositionSide_POSITION_SIDE_SHORT)
	}
	return 0
}

func closePositionSide(orderSide int64) int64 {
	if orderSide == int64(option.Side_SIDE_SELL) {
		return int64(option.PositionSide_POSITION_SIDE_LONG)
	}
	if orderSide == int64(option.Side_SIDE_BUY) {
		return int64(option.PositionSide_POSITION_SIDE_SHORT)
	}
	return 0
}

func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func normalizeZero(value float64) float64 {
	if math.Abs(value) <= optionFloatEpsilon {
		return 0
	}
	return value
}

func applyTradeToOrder(order *models.TOptionOrder, contract *models.TOptionContract, price, qty float64, now int64) {
	prevFilled := order.FilledQty
	nextFilled := prevFilled + qty
	if nextFilled <= optionFloatEpsilon {
		return
	}

	order.AvgPrice = ((order.AvgPrice * prevFilled) + (price * qty)) / nextFilled
	order.FilledQty = normalizeZero(nextFilled)
	order.UnfilledQty = normalizeZero(maxFloat64(order.Qty-order.FilledQty, 0))
	order.Turnover += optionTurnover(contract, price, qty)
	order.MatchTime = now
	order.UpdateTimes = now
	if order.UnfilledQty <= optionFloatEpsilon {
		order.UnfilledQty = 0
		order.Status = int64(option.OrderStatus_ORDER_STATUS_FILLED)
		return
	}
	order.Status = int64(option.OrderStatus_ORDER_STATUS_PART_FILLED)
}

func updateOpenPosition(ctx context.Context, model models.OptionPositionModel, contract *models.TOptionContract, order *models.TOptionOrder, price, qty float64, now int64) error {
	side := openPositionSide(order.Side)
	if side == 0 {
		return fmt.Errorf("invalid open order side: %d", order.Side)
	}

	pos, err := model.FindOneByTenantIdUserIdAccountIdContractIdSide(ctx, order.TenantId, order.UserId, order.AccountId, order.ContractId, side)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}

	multiplier := optionMultiplier(contract)
	if errors.Is(err, models.ErrNotFound) {
		exerciseableQty := 0.0
		if side == int64(option.PositionSide_POSITION_SIDE_LONG) {
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
			PositionValue:    price * qty * multiplier,
			ExerciseableQty:  exerciseableQty,
			Status:           int64(option.PositionStatus_POSITION_STATUS_HOLDING),
			LastCalcTime:     now,
			CreateTimes:      now,
			UpdateTimes:      now,
		})
		return err
	}

	nextQty := pos.PositionQty + qty
	if nextQty <= optionFloatEpsilon {
		return nil
	}
	pos.OpenAvgPrice = ((pos.OpenAvgPrice * pos.PositionQty) + (price * qty)) / nextQty
	pos.PositionQty = normalizeZero(nextQty)
	pos.AvailableQty = normalizeZero(pos.AvailableQty + qty)
	pos.MarkPrice = price
	pos.PositionValue = pos.MarkPrice * pos.PositionQty * multiplier
	if side == int64(option.PositionSide_POSITION_SIDE_LONG) {
		pos.ExerciseableQty = normalizeZero(pos.ExerciseableQty + qty)
		pos.UnrealizedPnl = (pos.MarkPrice - pos.OpenAvgPrice) * pos.PositionQty * multiplier
	} else {
		pos.UnrealizedPnl = (pos.OpenAvgPrice - pos.MarkPrice) * pos.PositionQty * multiplier
	}
	pos.Status = int64(option.PositionStatus_POSITION_STATUS_HOLDING)
	pos.LastCalcTime = now
	pos.UpdateTimes = now
	return model.Update(ctx, pos)
}

func updateClosePosition(ctx context.Context, model models.OptionPositionModel, contract *models.TOptionContract, order *models.TOptionOrder, price, qty float64, now int64) error {
	side := closePositionSide(order.Side)
	if side == 0 {
		return fmt.Errorf("invalid close order side: %d", order.Side)
	}

	pos, err := model.FindOneByTenantIdUserIdAccountIdContractIdSide(ctx, order.TenantId, order.UserId, order.AccountId, order.ContractId, side)
	if err != nil {
		return err
	}
	if pos.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) {
		return fmt.Errorf("position is not holding")
	}

	reduceQty := minFloat64(qty, pos.PositionQty)
	multiplier := optionMultiplier(contract)
	if side == int64(option.PositionSide_POSITION_SIDE_LONG) {
		pos.RealizedPnl += (price - pos.OpenAvgPrice) * reduceQty * multiplier
		pos.ExerciseableQty = normalizeZero(maxFloat64(pos.ExerciseableQty-reduceQty, 0))
	} else {
		pos.RealizedPnl += (pos.OpenAvgPrice - price) * reduceQty * multiplier
	}

	if pos.FrozenQty >= reduceQty {
		pos.FrozenQty = normalizeZero(pos.FrozenQty - reduceQty)
	} else {
		left := reduceQty - pos.FrozenQty
		pos.FrozenQty = 0
		pos.AvailableQty = normalizeZero(maxFloat64(pos.AvailableQty-left, 0))
	}
	pos.PositionQty = normalizeZero(maxFloat64(pos.PositionQty-reduceQty, 0))
	pos.MarkPrice = price
	pos.PositionValue = pos.MarkPrice * pos.PositionQty * multiplier
	pos.LastCalcTime = now
	pos.UpdateTimes = now
	if pos.PositionQty <= optionFloatEpsilon {
		pos.PositionQty = 0
		pos.AvailableQty = 0
		pos.FrozenQty = 0
		pos.ExerciseableQty = 0
		pos.UnrealizedPnl = 0
		pos.PositionValue = 0
		pos.Status = int64(option.PositionStatus_POSITION_STATUS_CLOSED)
	}
	return model.Update(ctx, pos)
}

func updatePositionByFilledOrder(ctx context.Context, model models.OptionPositionModel, contract *models.TOptionContract, order *models.TOptionOrder, price, qty float64, now int64) error {
	if order.PositionEffect == int64(option.PositionEffect_POSITION_EFFECT_CLOSE) {
		return updateClosePosition(ctx, model, contract, order, price, qty, now)
	}
	return updateOpenPosition(ctx, model, contract, order, price, qty, now)
}

func freezeClosePosition(ctx context.Context, model models.OptionPositionModel, order *models.TOptionOrder, now int64) error {
	if order.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_CLOSE) {
		return nil
	}

	side := closePositionSide(order.Side)
	if side == 0 {
		return fmt.Errorf("invalid close order side: %d", order.Side)
	}
	pos, err := model.FindOneByTenantIdUserIdAccountIdContractIdSide(ctx, order.TenantId, order.UserId, order.AccountId, order.ContractId, side)
	if err != nil {
		return err
	}
	if pos.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) || pos.AvailableQty+optionFloatEpsilon < order.Qty {
		return errInsufficientPositionQuantity
	}

	pos.AvailableQty = normalizeZero(pos.AvailableQty - order.Qty)
	pos.FrozenQty = normalizeZero(pos.FrozenQty + order.Qty)
	pos.UpdateTimes = now
	return model.Update(ctx, pos)
}

func releaseClosePositionFrozenQty(ctx context.Context, model models.OptionPositionModel, order *models.TOptionOrder, qty float64, now int64) error {
	if order.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_CLOSE) || qty <= optionFloatEpsilon {
		return nil
	}

	side := closePositionSide(order.Side)
	if side == 0 {
		return fmt.Errorf("invalid close order side: %d", order.Side)
	}
	pos, err := model.FindOneByTenantIdUserIdAccountIdContractIdSide(ctx, order.TenantId, order.UserId, order.AccountId, order.ContractId, side)
	if err != nil {
		return err
	}
	releaseQty := minFloat64(qty, pos.FrozenQty)
	pos.FrozenQty = normalizeZero(maxFloat64(pos.FrozenQty-releaseQty, 0))
	pos.AvailableQty = normalizeZero(pos.AvailableQty + releaseQty)
	pos.UpdateTimes = now
	return model.Update(ctx, pos)
}

func optionIntrinsicValue(contract *models.TOptionContract, deliveryPrice float64) float64 {
	if contract.OptionType == int64(option.OptionType_OPTION_TYPE_CALL) {
		return maxFloat64(deliveryPrice-contract.StrikePrice, 0)
	}
	if contract.OptionType == int64(option.OptionType_OPTION_TYPE_PUT) {
		return maxFloat64(contract.StrikePrice-deliveryPrice, 0)
	}
	return 0
}

func optionSettlementPayoff(contract *models.TOptionContract, deliveryPrice, qty float64) float64 {
	return optionIntrinsicValue(contract, deliveryPrice) * qty * optionMultiplier(contract)
}

func optionExerciseAmount(contract *models.TOptionContract, qty float64) float64 {
	return contract.StrikePrice * qty * optionMultiplier(contract)
}

func applyOptionAccountDelta(ctx context.Context, accountModel models.OptionAccountModel, billModel models.OptionBillModel, tenantId, userId, accountId int64, coin string, amount float64, refType, refId int64, bizNo, remark string, realized bool, now int64) error {
	if math.Abs(amount) <= optionFloatEpsilon {
		return nil
	}

	account, err := accountModel.FindOneByTenantIdUserIdAccountIdMarginCoin(ctx, tenantId, userId, accountId, coin)
	before := 0.0
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
		account.Balance = normalizeZero(account.Balance + amount)
		account.AvailableBalance = normalizeZero(account.AvailableBalance + amount)
		if realized {
			account.RealizedPnl = normalizeZero(account.RealizedPnl + amount)
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
		BalanceAfter:  before + amount,
		Remark:        remark,
		CreateTimes:   now,
	})
	return err
}
