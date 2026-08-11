package helpers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/domain/contractmath"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

// UserTradeControlDecision is the single runtime decision produced from the
// product-level and symbol-level user controls.
type UserTradeControlDecision struct {
	Passed     bool
	RejectCode string
	RejectMsg  string
}

type userTradeControlSnapshot struct {
	ProductLimitID int64 `json:"product_limit_id,omitempty"`
	SymbolLimitID  int64 `json:"symbol_limit_id,omitempty"`
	ControlMode    int64 `json:"control_mode"`
}

// EvaluateAndLogUserOrderRisk is shared by App, Internal and Task RPC
// implementations so all entry points enforce exactly the same policy.
func EvaluateAndLogUserOrderRisk(
	ctx context.Context, svcCtx *svc.ServiceContext, in *trade.CheckOrderRiskReq,
) (*trade.CheckOrderRiskResp, error) {
	symbol, err := svcCtx.TradeSymbolModel.FindOne(ctx, in.SymbolId)
	if err != nil {
		return nil, err
	}
	if symbol.TenantId != in.TenantId {
		return nil, models.ErrNotFound
	}

	productLimit, err := findEffectiveProductLimit(ctx, svcCtx, in, symbol)
	if err != nil {
		return nil, err
	}
	symbolLimit, err := findEffectiveSymbolLimit(ctx, svcCtx, in)
	if err != nil {
		return nil, err
	}
	decision, err := evaluateUserOrderRisk(ctx, svcCtx, in, symbol, productLimit, symbolLimit)
	if err != nil {
		return nil, err
	}

	mode := effectiveControlMode(productLimit, symbolLimit)
	snapshot, _ := json.Marshal(userTradeControlSnapshot{
		ProductLimitID: idOfProductLimit(productLimit),
		SymbolLimitID:  idOfSymbolLimit(symbolLimit),
		ControlMode:    mode,
	})
	checkResult := trade.RiskCheckResult_RISK_CHECK_RESULT_PASS
	if !decision.Passed {
		checkResult = trade.RiskCheckResult_RISK_CHECK_RESULT_REJECT
	}
	qty := MustParseFloat(in.Qty)
	amount := MustParseFloat(in.Amount)
	if _, err = svcCtx.RiskOrderCheckLogModel.Insert(ctx, &models.TRiskOrderCheckLog{
		TenantId:      in.TenantId,
		UserId:        in.UserId,
		SymbolId:      in.SymbolId,
		ProductType:   symbol.ProductType,
		CheckType:     int64(trade.RiskCheckType_RISK_CHECK_TYPE_TRADE_PERMISSION),
		CheckResult:   int64(checkResult),
		RejectCode:    decision.RejectCode,
		RejectMsg:     decision.RejectMsg,
		RequestPrice:  MustParseFloat(in.Price),
		RequestQty:    qty,
		RequestAmount: amount,
		OperatorId:    in.UserId,
		Source:        int64(trade.SourceType_SOURCE_TYPE_USER),
		CheckSnapshot: sql.NullString{String: string(snapshot), Valid: len(snapshot) > 0},
		CreateTimes:   utils.NowMillis(),
	}); err != nil {
		return nil, err
	}
	return &trade.CheckOrderRiskResp{
		Passed:     boolToInt64(decision.Passed),
		RejectCode: decision.RejectCode,
		RejectMsg:  decision.RejectMsg,
	}, nil
}

// CheckUserCancelAllowed enforces the same effective product control for
// single-order and bulk cancellation paths.
func CheckUserCancelAllowed(
	ctx context.Context, svcCtx *svc.ServiceContext, order *models.TTradeOrder,
) error {
	if order == nil {
		return models.ErrNotFound
	}
	in := &trade.CheckOrderRiskReq{
		TenantId: order.TenantId,
		UserId:   order.UserId,
		SymbolId: order.SymbolId,
	}
	symbol, err := svcCtx.TradeSymbolModel.FindOne(ctx, order.SymbolId)
	if err != nil {
		return err
	}
	if symbol.TenantId != order.TenantId {
		return models.ErrNotFound
	}
	limit, err := findEffectiveProductLimit(ctx, svcCtx, in, symbol)
	if err != nil {
		return err
	}
	if limit == nil {
		return nil
	}
	if limit.CanCancel == 0 {
		return errors.New("user cancellation is disabled")
	}
	if limit.MaxCancelCountPerDay > 0 {
		count, countErr := svcCtx.TradeOrderModel.CountCancelsSince(
			ctx, order.TenantId, order.UserId, order.ProductType,
			productContractType(order.ProductType, order.ContractType), startOfUTCDayMillis(),
		)
		if countErr != nil {
			return countErr
		}
		if count >= limit.MaxCancelCountPerDay {
			return errors.New("daily cancellation limit reached")
		}
	}
	return nil
}

func evaluateUserOrderRisk(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	in *trade.CheckOrderRiskReq,
	symbol *models.TTradeSymbol,
	productLimit *models.TRiskUserTradeLimit,
	symbolLimit *models.TRiskUserSymbolLimit,
) (UserTradeControlDecision, error) {
	pass := UserTradeControlDecision{Passed: true}
	reject := func(code, msg string) UserTradeControlDecision {
		return UserTradeControlDecision{RejectCode: code, RejectMsg: msg}
	}
	increasing := in.ExposureIncreasing == common.YesNo_YES_NO_YES
	mode := effectiveControlMode(productLimit, symbolLimit)
	switch mode {
	case int64(trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_DISABLED):
		return reject("TRADE_DISABLED", "user trading is disabled"), nil
	case int64(trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_CLOSE_ONLY):
		if increasing {
			return reject("REDUCE_ONLY", "only exposure-reducing orders are allowed"), nil
		}
	case int64(trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_REDUCE_ONLY):
		if increasing || (symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) && in.IsReduceOnly != common.YesNo_YES_NO_YES) {
			return reject("REDUCE_ONLY", "an explicit exposure-reducing order is required"), nil
		}
	}

	if productLimit != nil {
		if productLimit.TradeEnabled == int64(common.Enable_ENABLE_DISABLED) {
			return reject("TRADE_DISABLED", "user trading is disabled"), nil
		}
		if increasing && productLimit.CanOpen == 0 {
			return reject("OPEN_DISABLED", "opening or increasing exposure is disabled"), nil
		}
		if !increasing && productLimit.CanClose == 0 {
			return reject("CLOSE_DISABLED", "closing or reducing exposure is disabled"), nil
		}
		if increasing && productLimit.OnlyReduceOnly == int64(common.Enable_ENABLE_ENABLED) {
			return reject("REDUCE_ONLY", "only exposure-reducing orders are allowed"), nil
		}
		if in.TriggerKind != trade.TriggerKind_TRIGGER_KIND_NONE && productLimit.CanTriggerOrder == 0 {
			return reject("TRIGGER_ORDER_DISABLED", "trigger orders are disabled"), nil
		}
		if in.OrderSource == trade.OrderSourceType_ORDER_SOURCE_TYPE_API && productLimit.CanApiTrade == 0 {
			return reject("API_TRADE_DISABLED", "API trading is disabled"), nil
		}
		if increasing && productLimit.MaxOpenNotional.IsPositive() && MustParseFloat(in.Amount).GreaterThan(productLimit.MaxOpenNotional) {
			return reject("MAX_OPEN_NOTIONAL", "opening notional exceeds the user limit"), nil
		}
		if productLimit.MaxOpenOrderCount > 0 {
			count, err := svcCtx.TradeOrderModel.CountByStatuses(ctx, uint64(in.TenantId), uint64(in.UserId), symbol.ProductType, OpenOrderStatuses())
			if err != nil {
				return pass, err
			}
			if count >= productLimit.MaxOpenOrderCount {
				return reject("MAX_OPEN_ORDERS", "maximum open order count reached"), nil
			}
		}
		if productLimit.MaxOrderCountPerDay > 0 {
			count, err := svcCtx.TradeOrderModel.CountCreatedSince(
				ctx, in.TenantId, in.UserId, symbol.ProductType,
				productContractType(symbol.ProductType, symbol.ContractType), startOfUTCDayMillis(),
			)
			if err != nil {
				return pass, err
			}
			if count >= productLimit.MaxOrderCountPerDay {
				return reject("DAILY_ORDER_LIMIT", "daily order limit reached"), nil
			}
		}
	}

	if symbolLimit != nil {
		qty := MustParseFloat(in.Qty)
		notional := MustParseFloat(in.Amount)
		if symbolLimit.MinOrderQty.IsPositive() && qty.IsPositive() && qty.LessThan(symbolLimit.MinOrderQty) {
			return reject("MIN_QTY", "quantity below the user limit"), nil
		}
		if symbolLimit.MaxOrderQty.IsPositive() && qty.IsPositive() && qty.GreaterThan(symbolLimit.MaxOrderQty) {
			return reject("MAX_QTY", "quantity exceeds the user limit"), nil
		}
		if symbolLimit.MinOrderNotional.IsPositive() && notional.IsPositive() && notional.LessThan(symbolLimit.MinOrderNotional) {
			return reject("MIN_NOTIONAL", "notional below the user limit"), nil
		}
		if symbolLimit.MaxOrderNotional.IsPositive() && notional.IsPositive() && notional.GreaterThan(symbolLimit.MaxOrderNotional) {
			return reject("MAX_NOTIONAL", "notional exceeds the user limit"), nil
		}
		if symbolLimit.MaxOpenOrders > 0 {
			count, err := svcCtx.TradeOrderModel.CountByUserSymbolStatuses(ctx, in.TenantId, in.UserId, in.SymbolId, OpenOrderStatuses())
			if err != nil {
				return pass, err
			}
			if count >= symbolLimit.MaxOpenOrders {
				return reject("MAX_SYMBOL_OPEN_ORDERS", "maximum symbol open order count reached"), nil
			}
		}
		if symbolLimit.PriceDeviationRate.IsPositive() && in.OrderType == trade.OrderType_ORDER_TYPE_LIMIT {
			orderPrice := MustParseFloat(in.Price)
			referencePrice := MustParseFloat(in.ReferencePrice)
			if !referencePrice.IsPositive() {
				return reject("REFERENCE_PRICE_UNAVAILABLE", "a current reference price is required"), nil
			}
			deviation := orderPrice.Sub(referencePrice).Abs().Div(referencePrice)
			if deviation.GreaterThan(symbolLimit.PriceDeviationRate) {
				return reject("PRICE_DEVIATION", "order price deviation exceeds the user limit"), nil
			}
		}
	}
	if increasing && symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		if decision, err := evaluateDerivativePositionLimits(ctx, svcCtx, in, symbol, productLimit, symbolLimit); err != nil || !decision.Passed {
			return decision, err
		}
	}
	if increasing && symbol.ProductType == int64(common.ProductType_PRODUCT_TYPE_SPOT) {
		if decision, err := evaluateSpotPositionLimits(ctx, svcCtx, in, symbol, symbolLimit); err != nil || !decision.Passed {
			return decision, err
		}
	}
	return pass, nil
}

func evaluateSpotPositionLimits(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	in *trade.CheckOrderRiskReq,
	symbol *models.TTradeSymbol,
	symbolLimit *models.TRiskUserSymbolLimit,
) (UserTradeControlDecision, error) {
	pass := UserTradeControlDecision{Passed: true}
	if symbolLimit == nil || (!symbolLimit.MaxPositionQty.IsPositive() && !symbolLimit.MaxPositionNotional.IsPositive()) {
		return pass, nil
	}
	resp, err := svcCtx.AssetClient.GetAssetBalance(ctx, &asset.GetUserAssetDetailReq{
		TenantId:   in.TenantId,
		UserId:     in.UserId,
		WalletType: WalletTypeForTrade(common.ProductType(symbol.ProductType), symbol.CategoryType),
		Coin:       symbol.BaseAsset,
	})
	if err != nil {
		return pass, err
	}
	currentQty := decimal.Zero
	if resp != nil && resp.GetBase() != nil && resp.GetBase().GetCode() == i18n.AssetNotFound {
		// A user with no base-asset row has a zero position.
	} else if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		return pass, fmt.Errorf("get spot asset balance rejected, tenantId=%d userId=%d coin=%s", in.TenantId, in.UserId, symbol.BaseAsset)
	} else {
		currentQty = MustParseFloat(resp.GetData().GetTotalAmount())
	}
	return evaluateSpotPositionExposure(
		currentQty,
		MustParseFloat(in.Qty),
		MustParseFloat(in.Amount),
		MustParseFloat(in.ReferencePrice),
		MustParseFloat(in.Price),
		symbolLimit.MaxPositionQty,
		symbolLimit.MaxPositionNotional,
	), nil
}

func evaluateSpotPositionExposure(
	currentQty, orderQty, orderNotional, referencePrice, orderPrice, maxQty, maxNotional decimal.Decimal,
) UserTradeControlDecision {
	pass := UserTradeControlDecision{Passed: true}
	nextQty := currentQty.Add(orderQty)
	if maxQty.IsPositive() && nextQty.GreaterThan(maxQty) {
		return UserTradeControlDecision{RejectCode: "MAX_POSITION_QTY", RejectMsg: "spot holding quantity exceeds the user limit"}
	}
	if !maxNotional.IsPositive() {
		return pass
	}
	price := referencePrice
	if !price.IsPositive() {
		price = orderPrice
	}
	if !price.IsPositive() {
		return UserTradeControlDecision{RejectCode: "REFERENCE_PRICE_UNAVAILABLE", RejectMsg: "a current reference price is required"}
	}
	if !orderNotional.IsPositive() {
		orderNotional = orderQty.Mul(price)
	}
	if currentQty.Mul(price).Add(orderNotional).GreaterThan(maxNotional) {
		return UserTradeControlDecision{RejectCode: "MAX_POSITION_NOTIONAL", RejectMsg: "spot holding notional exceeds the user limit"}
	}
	return pass
}

type positionExposure struct {
	qty      decimal.Decimal
	longQty  decimal.Decimal
	shortQty decimal.Decimal
	notional decimal.Decimal
}

func evaluateDerivativePositionLimits(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	in *trade.CheckOrderRiskReq,
	symbol *models.TTradeSymbol,
	productLimit *models.TRiskUserTradeLimit,
	symbolLimit *models.TRiskUserSymbolLimit,
) (UserTradeControlDecision, error) {
	pass := UserTradeControlDecision{Passed: true}
	reject := func(code, msg string) UserTradeControlDecision {
		return UserTradeControlDecision{RejectCode: code, RejectMsg: msg}
	}
	orderQty := MustParseFloat(in.Qty)
	orderNotional := MustParseFloat(in.Amount)

	if symbolLimit != nil && (symbolLimit.MaxPositionQty.IsPositive() || symbolLimit.MaxLongPositionQty.IsPositive() || symbolLimit.MaxShortPositionQty.IsPositive() || symbolLimit.MaxPositionNotional.IsPositive()) {
		exposure, err := loadDerivativePositionExposure(ctx, svcCtx, in.TenantId, in.UserId, symbol.Id, symbol.ContractType)
		if err != nil {
			return pass, err
		}
		if symbolLimit.MaxPositionQty.IsPositive() && exposure.qty.Add(orderQty).GreaterThan(symbolLimit.MaxPositionQty) {
			return reject("MAX_POSITION_QTY", "position quantity exceeds the user limit"), nil
		}
		if symbolLimit.MaxPositionNotional.IsPositive() && exposure.notional.Add(orderNotional).GreaterThan(symbolLimit.MaxPositionNotional) {
			return reject("MAX_POSITION_NOTIONAL", "position notional exceeds the symbol limit"), nil
		}
		longIncrease, shortIncrease := derivativeDirectionIncrease(in, orderQty)
		if symbolLimit.MaxLongPositionQty.IsPositive() && exposure.longQty.Add(longIncrease).GreaterThan(symbolLimit.MaxLongPositionQty) {
			return reject("MAX_LONG_POSITION_QTY", "long position quantity exceeds the user limit"), nil
		}
		if symbolLimit.MaxShortPositionQty.IsPositive() && exposure.shortQty.Add(shortIncrease).GreaterThan(symbolLimit.MaxShortPositionQty) {
			return reject("MAX_SHORT_POSITION_QTY", "short position quantity exceeds the user limit"), nil
		}
	}
	if productLimit != nil && productLimit.MaxPositionNotional.IsPositive() {
		exposure, err := loadDerivativePositionExposure(ctx, svcCtx, in.TenantId, in.UserId, 0, productContractType(symbol.ProductType, symbol.ContractType))
		if err != nil {
			return pass, err
		}
		if exposure.notional.Add(orderNotional).GreaterThan(productLimit.MaxPositionNotional) {
			return reject("MAX_PRODUCT_POSITION_NOTIONAL", "position notional exceeds the product limit"), nil
		}
	}
	return pass, nil
}

func loadDerivativePositionExposure(
	ctx context.Context, svcCtx *svc.ServiceContext, tenantID, userID, symbolID, contractType int64,
) (positionExposure, error) {
	result := positionExposure{}
	positions, err := svcCtx.ContractPositionModel.FindList(ctx, models.ContractPositionPageFilter{
		TenantId: tenantID, UserId: userID, SymbolId: symbolID, ContractType: contractType,
	})
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return result, err
	}
	for _, position := range positions {
		if position == nil || position.Status != 1 || !position.Qty.IsPositive() {
			continue
		}
		result.qty = result.qty.Add(position.Qty)
		switch trade.PositionSide(position.PositionSide) {
		case trade.PositionSide_POSITION_SIDE_LONG:
			result.longQty = result.longQty.Add(position.Qty)
		case trade.PositionSide_POSITION_SIDE_SHORT:
			result.shortQty = result.shortQty.Add(position.Qty)
		}
		positionSymbol, findErr := svcCtx.TradeSymbolModel.FindOne(ctx, position.SymbolId)
		if findErr != nil {
			return result, findErr
		}
		contract, findErr := svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(ctx, tenantID, position.SymbolId)
		if findErr != nil {
			return result, findErr
		}
		price := position.MarkPrice
		if !price.IsPositive() {
			price = position.OpenAvgPrice
		}
		values, valueErr := contractmath.CalculateTradeValues(positionSymbol.ContractValueType, position.Qty, contract.ContractSize, price)
		if valueErr != nil {
			return result, valueErr
		}
		result.notional = result.notional.Add(values.QuoteNotional)
	}
	return result, nil
}

func derivativeDirectionIncrease(in *trade.CheckOrderRiskReq, qty decimal.Decimal) (decimal.Decimal, decimal.Decimal) {
	switch in.PositionSide {
	case trade.PositionSide_POSITION_SIDE_LONG:
		return qty, decimal.Zero
	case trade.PositionSide_POSITION_SIDE_SHORT:
		return decimal.Zero, qty
	default:
		if in.Side == common.Side_SIDE_BUY {
			return qty, decimal.Zero
		}
		return decimal.Zero, qty
	}
}

func findEffectiveProductLimit(
	ctx context.Context, svcCtx *svc.ServiceContext, in *trade.CheckOrderRiskReq, symbol *models.TTradeSymbol,
) (*models.TRiskUserTradeLimit, error) {
	contractType := productContractType(symbol.ProductType, symbol.ContractType)
	item, err := svcCtx.RiskUserTradeLimitModel.FindOneByTenantIdUserIdProductTypeContractType(
		ctx, in.TenantId, in.UserId, symbol.ProductType, contractType,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil && contractType > 0 {
		item, err = svcCtx.RiskUserTradeLimitModel.FindOneByTenantIdUserIdProductTypeContractType(
			ctx, in.TenantId, in.UserId, symbol.ProductType, 0,
		)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
	}
	if !isProductLimitEffective(item, utils.NowMillis()) {
		return nil, nil
	}
	return item, nil
}

func findEffectiveSymbolLimit(
	ctx context.Context, svcCtx *svc.ServiceContext, in *trade.CheckOrderRiskReq,
) (*models.TRiskUserSymbolLimit, error) {
	item, err := svcCtx.RiskUserSymbolLimitModel.FindOneByTenantIdUserIdSymbolId(ctx, in.TenantId, in.UserId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if !isSymbolLimitEffective(item, utils.NowMillis()) {
		return nil, nil
	}
	return item, nil
}

func isProductLimitEffective(item *models.TRiskUserTradeLimit, now int64) bool {
	return item != nil && item.Enabled == int64(common.Enable_ENABLE_ENABLED) &&
		(item.EffectiveStartTime == 0 || now >= item.EffectiveStartTime) &&
		(item.EffectiveEndTime == 0 || now < item.EffectiveEndTime)
}

func isSymbolLimitEffective(item *models.TRiskUserSymbolLimit, now int64) bool {
	return item != nil && item.Enabled == int64(common.Enable_ENABLE_ENABLED) &&
		(item.EffectiveStartTime == 0 || now >= item.EffectiveStartTime) &&
		(item.EffectiveEndTime == 0 || now < item.EffectiveEndTime)
}

func effectiveControlMode(product *models.TRiskUserTradeLimit, symbol *models.TRiskUserSymbolLimit) int64 {
	mode := int64(trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_NORMAL)
	if product != nil && product.ControlMode > mode {
		mode = product.ControlMode
	}
	if symbol != nil && symbol.ControlMode > mode {
		mode = symbol.ControlMode
	}
	return mode
}

func productContractType(productType, contractType int64) int64 {
	if productType == int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE) {
		return contractType
	}
	return 0
}

func startOfUTCDayMillis() int64 {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).UnixMilli()
}

func idOfProductLimit(item *models.TRiskUserTradeLimit) int64 {
	if item == nil {
		return 0
	}
	return item.Id
}

func idOfSymbolLimit(item *models.TRiskUserSymbolLimit) int64 {
	if item == nil {
		return 0
	}
	return item.Id
}

func boolToInt64(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func YesNo(value bool) common.YesNo {
	if value {
		return common.YesNo_YES_NO_YES
	}
	return common.YesNo_YES_NO_NO
}
