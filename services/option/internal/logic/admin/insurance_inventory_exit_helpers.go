package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/metadata"
)

var errInvalidInsuranceInventoryExit = errors.New("invalid insurance inventory exit")
var errOpenInsuranceInventoryExit = errors.New("an unsubmitted insurance inventory exit already exists")
var errInsuranceInventoryExitDisabled = errors.New("insurance inventory exit is disabled")
var errInsuranceInventoryExitLimits = errors.New("insurance inventory exit runtime limits are invalid")

type insuranceInventoryExitLimits struct {
	maxQuantityPerRequest decimal.Decimal
	maxPremiumPerRequest  decimal.Decimal
	maxDailyQuantity      decimal.Decimal
	maxMarkDeviationRatio decimal.Decimal
	minOrderBookQuantity  decimal.Decimal
}

func insuranceInventoryExitRuntimeLimits(svcCtx *svc.ServiceContext) (insuranceInventoryExitLimits, error) {
	if svcCtx == nil || !svcCtx.Config.InsuranceInventoryExit.Enabled {
		return insuranceInventoryExitLimits{}, errInsuranceInventoryExitDisabled
	}
	parse := func(value string) (decimal.Decimal, error) {
		parsed, err := decimal.NewFromString(strings.TrimSpace(value))
		if err != nil || !parsed.IsPositive() {
			return decimal.Zero, errInsuranceInventoryExitLimits
		}
		return parsed, nil
	}
	config := svcCtx.Config.InsuranceInventoryExit
	maxQuantity, err := parse(config.MaxQuantityPerRequest)
	if err != nil {
		return insuranceInventoryExitLimits{}, err
	}
	maxPremium, err := parse(config.MaxPremiumPerRequest)
	if err != nil {
		return insuranceInventoryExitLimits{}, err
	}
	maxDaily, err := parse(config.MaxDailyQuantity)
	if err != nil {
		return insuranceInventoryExitLimits{}, err
	}
	maxDeviation, err := parse(config.MaxMarkDeviationRatio)
	if err != nil || maxDeviation.GreaterThan(decimal.NewFromInt(1)) {
		return insuranceInventoryExitLimits{}, errInsuranceInventoryExitLimits
	}
	minDepth, err := parse(config.MinOrderBookQuantity)
	if err != nil || maxQuantity.GreaterThan(maxDaily) {
		return insuranceInventoryExitLimits{}, errInsuranceInventoryExitLimits
	}
	return insuranceInventoryExitLimits{
		maxQuantityPerRequest: maxQuantity,
		maxPremiumPerRequest:  maxPremium,
		maxDailyQuantity:      maxDaily,
		maxMarkDeviationRatio: maxDeviation,
		minOrderBookQuantity:  minDepth,
	}, nil
}

func insuranceInventoryExitError(ctx context.Context, err error) *option.GetInsuranceInventoryExitResp {
	message := i18n.Translate(i18n.OperationNotAllowed, ctx)
	if err != nil {
		message = err.Error()
	}
	return &option.GetInsuranceInventoryExitResp{
		Base: helper.ErrResp(i18n.OperationNotAllowed, message),
	}
}

func validateInsuranceInventoryExit(
	position *models.TOptionPosition,
	contract *models.TOptionContract,
	market *models.TOptionMarket,
	quantity, limitPrice decimal.Decimal,
	now int64,
) error {
	if position == nil || contract == nil || market == nil ||
		position.TenantId != contract.TenantId || position.ContractId != contract.Id ||
		position.Side != int64(common.PositionSide_POSITION_SIDE_SHORT) ||
		position.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) ||
		contract.InsuranceUserId <= 0 || contract.InsuranceAccountId <= 0 ||
		position.UserId != contract.InsuranceUserId ||
		position.AccountId != contract.InsuranceAccountId ||
		contract.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) ||
		contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) ||
		contract.IsDeleted == int64(common.YesNo_YES_NO_YES) || now < contract.ListTime ||
		(contract.LastTradeTime <= 0 || now >= contract.LastTradeTime) ||
		!quantity.IsPositive() || position.AvailableQty.LessThan(quantity) ||
		!limitPrice.IsPositive() || !logichelpers.IsMarkFresh(market, now, 30) ||
		!market.MarkPrice.IsPositive() || !contract.OrderPriceBandRatio.IsPositive() ||
		contract.OrderPriceBandRatio.GreaterThan(decimal.NewFromInt(1)) {
		return errInvalidInsuranceInventoryExit
	}
	if contract.MinOrderQty.IsPositive() && quantity.LessThan(contract.MinOrderQty) {
		return errInvalidInsuranceInventoryExit
	}
	if contract.MaxOrderQty.IsPositive() && quantity.GreaterThan(contract.MaxOrderQty) {
		return errInvalidInsuranceInventoryExit
	}
	if !contract.QtyStep.IsPositive() || !quantity.Mod(contract.QtyStep).IsZero() ||
		!contract.PriceTick.IsPositive() || !limitPrice.Mod(contract.PriceTick).IsZero() {
		return errInvalidInsuranceInventoryExit
	}
	lower := decimal.Max(
		market.MarkPrice.Mul(decimal.NewFromInt(1).Sub(contract.OrderPriceBandRatio)),
		decimal.Zero,
	)
	upper := market.MarkPrice.Mul(decimal.NewFromInt(1).Add(contract.OrderPriceBandRatio))
	if limitPrice.LessThan(lower) || limitPrice.GreaterThan(upper) {
		return fmt.Errorf("%w: limit price is outside the current mark-price band", errInvalidInsuranceInventoryExit)
	}
	return nil
}

func validateInsuranceInventoryExitRuntimeLimits(
	contract *models.TOptionContract,
	market *models.TOptionMarket,
	quantity, limitPrice decimal.Decimal,
	limits insuranceInventoryExitLimits,
) error {
	if contract == nil || market == nil || !contract.Multiplier.IsPositive() ||
		!market.MarkPrice.IsPositive() || quantity.GreaterThan(limits.maxQuantityPerRequest) {
		return errInsuranceInventoryExitLimits
	}
	premium := quantity.Mul(limitPrice).Mul(contract.Multiplier)
	if premium.GreaterThan(limits.maxPremiumPerRequest) {
		return fmt.Errorf("%w: premium limit exceeded", errInsuranceInventoryExitLimits)
	}
	deviation := limitPrice.Sub(market.MarkPrice).Abs().Div(market.MarkPrice)
	if deviation.GreaterThan(limits.maxMarkDeviationRatio) {
		return fmt.Errorf("%w: mark-price deviation limit exceeded", errInsuranceInventoryExitLimits)
	}
	return nil
}

func validateInsuranceInventoryExitOrderBookDepth(
	orders []*models.TOptionOrder,
	limits insuranceInventoryExitLimits,
) error {
	depth := decimal.Zero
	for _, order := range orders {
		if order != nil && order.UnfilledQty.IsPositive() {
			depth = depth.Add(order.UnfilledQty)
		}
	}
	if depth.LessThan(limits.minOrderBookQuantity) {
		return fmt.Errorf("%w: executable order-book depth is below minimum", errInsuranceInventoryExitLimits)
	}
	return nil
}

func insuranceInventoryExitUTCDayStart(now time.Time) int64 {
	utc := now.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).Unix()
}

func insuranceInventoryExitUserContext(
	ctx context.Context, tenantID, insuranceUserID int64,
) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	md = md.Copy()
	md.Set(utils.CtxKeyUid, strconv.FormatInt(insuranceUserID, 10))
	md.Set(utils.CtxKeyTenantId, strconv.FormatInt(tenantID, 10))
	return metadata.NewIncomingContext(ctx, md)
}

func insuranceInventoryExitClientOrderID(exitID int64) string {
	return fmt.Sprintf("INS-EXIT-%d", exitID)
}

func saveInsuranceInventoryExitFailure(
	ctx context.Context, svcCtx *svc.ServiceContext, exitID int64, executionErr error,
) {
	if executionErr == nil {
		return
	}
	item, err := svcCtx.OptionInsuranceInventoryExitModel.FindOne(ctx, exitID)
	if err != nil || item.Status != int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_APPROVED) {
		return
	}
	item.LastErrorMsg = strings.TrimSpace(executionErr.Error())
	if len(item.LastErrorMsg) > 500 {
		item.LastErrorMsg = item.LastErrorMsg[:500]
	}
	item.UpdateTimes = time.Now().Unix()
	_ = svcCtx.OptionInsuranceInventoryExitModel.Update(ctx, item)
}
