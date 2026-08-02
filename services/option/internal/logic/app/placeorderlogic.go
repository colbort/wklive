package applogic

import (
	"context"
	"errors"
	"time"
	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type PlaceOrderLogic struct {
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	orderSource option.OrderSource
	logx.Logger
}

func NewPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceOrderLogic {
	return &PlaceOrderLogic{
		ctx:         ctx,
		svcCtx:      svcCtx,
		orderSource: option.OrderSource_ORDER_SOURCE_APP,
		Logger:      logx.WithContext(ctx),
	}
}

// NewAdministrativePlaceOrderLogic is reserved for governed internal workflows.
// Public callers cannot select ORDER_SOURCE_ADMIN through PlaceOrderReq.
func NewAdministrativePlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceOrderLogic {
	logic := NewPlaceOrderLogic(ctx, svcCtx)
	logic.orderSource = option.OrderSource_ORDER_SOURCE_ADMIN
	return logic
}

// 提交期权下单请求
func (l *PlaceOrderLogic) PlaceOrder(in *option.PlaceOrderReq) (*option.PlaceOrderResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if contract.TenantId != tenantId {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
	}
	if contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotTradable, i18n.Translate(i18n.ContractNotTradable, l.ctx))}, nil
	}
	entryNow := time.Now().Unix()
	if contract.IsDeleted == int64(common.YesNo_YES_NO_YES) ||
		entryNow < contract.ListTime ||
		(contract.LastTradeTime <= 0 || entryNow >= contract.LastTradeTime) {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotTradable, i18n.Translate(i18n.ContractNotTradable, l.ctx))}, nil
	}
	calendarDecision, calendarErr := logichelpers.IsContractTradingOpen(l.ctx, l.svcCtx, contract, entryNow)
	if calendarErr != nil || calendarDecision == nil || !calendarDecision.Open {
		reason := "CALENDAR_EVALUATION_FAILED"
		if calendarDecision != nil && calendarDecision.Reason != "" {
			reason = calendarDecision.Reason
		}
		l.Errorf(
			"option order entry denied by trading calendar, tenantId=%d contractId=%d reason=%s err=%v",
			tenantId, contract.Id, reason, calendarErr,
		)
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotTradable, i18n.Translate(i18n.ContractNotTradable, l.ctx))}, nil
	}
	riskAccount, riskErr := l.svcCtx.OptionRiskAccountModel.FindOneByTenantIdUserIdAccountIdSettleCoin(
		l.ctx, tenantId, userId, 0, contract.SettleCoin,
	)
	if riskErr != nil && !errors.Is(riskErr, models.ErrNotFound) {
		return nil, riskErr
	}
	if riskErr == nil {
		switch option.RiskAccountStatus(riskAccount.Status) {
		case option.RiskAccountStatus_RISK_ACCOUNT_STATUS_LIQUIDATING,
			option.RiskAccountStatus_RISK_ACCOUNT_STATUS_BANKRUPT,
			option.RiskAccountStatus_RISK_ACCOUNT_STATUS_RESTRICTED:
			if in.PositionEffect != option.PositionEffect_POSITION_EFFECT_CLOSE {
				return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
			}
		case option.RiskAccountStatus_RISK_ACCOUNT_STATUS_MARGIN_CALL:
			if in.PositionEffect != option.PositionEffect_POSITION_EFFECT_CLOSE {
				return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
			}
		}
	}
	if in.Side != common.Side_SIDE_BUY && in.Side != common.Side_SIDE_SELL {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if in.PositionEffect != option.PositionEffect_POSITION_EFFECT_OPEN &&
		in.PositionEffect != option.PositionEffect_POSITION_EFFECT_CLOSE {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if ready, reason := logichelpers.OrderProductScopeReady(
		contract, l.svcCtx.Config.ProductScope, in.Side, in.PositionEffect, in.Mmp,
	); !ready {
		l.Errorf(
			"option order denied by product scope tenantId=%d userId=%d contractId=%d reason=%s",
			tenantId, userId, contract.Id, reason,
		)
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, reason)}, nil
	}
	switch in.OrderType {
	case option.OrderType_ORDER_TYPE_LIMIT,
		option.OrderType_ORDER_TYPE_MARKET,
		option.OrderType_ORDER_TYPE_POST_ONLY,
		option.OrderType_ORDER_TYPE_IOC,
		option.OrderType_ORDER_TYPE_FOK:
	default:
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
	}
	if in.ReduceOnly == common.YesNo_YES_NO_YES && in.PositionEffect != option.PositionEffect_POSITION_EFFECT_CLOSE {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
	}
	mmpGroup := ""
	if in.Mmp == common.YesNo_YES_NO_YES {
		if in.OrderType != option.OrderType_ORDER_TYPE_POST_ONLY {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, controlReasonMMPInvalidOrder)}, nil
		}
		var valid bool
		mmpGroup, valid = NormalizeMMPGroup(in.MmpGroup)
		if !valid {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, controlReasonMMPInvalidOrder)}, nil
		}
	} else if in.MmpGroup != "" {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, controlReasonMMPInvalidOrder)}, nil
	}

	priceField := in.Price
	if in.OrderType == option.OrderType_ORDER_TYPE_MARKET {
		if in.Price != "" || in.ProtectionPrice == "" {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.PriceFormatError, i18n.Translate(i18n.PriceFormatError, l.ctx))}, nil
		}
		priceField = in.ProtectionPrice
	} else if in.ProtectionPrice != "" || in.MaxTurnover != "" {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	price, err := conv.ParseDecimalField(priceField)
	if err != nil || !price.IsPositive() {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.PriceFormatError, i18n.Translate(i18n.PriceFormatError, l.ctx))}, nil
	}
	qty, err := conv.ParseDecimalField(in.Qty)
	if err != nil || !qty.IsPositive() {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.QuantityFormatError, i18n.Translate(i18n.QuantityFormatError, l.ctx))}, nil
	}
	if contract.MinOrderQty.IsPositive() && qty.LessThan(contract.MinOrderQty) {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.QuantityFormatError, i18n.Translate(i18n.QuantityFormatError, l.ctx))}, nil
	}
	if contract.MaxOrderQty.IsPositive() && qty.GreaterThan(contract.MaxOrderQty) {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.QuantityFormatError, i18n.Translate(i18n.QuantityFormatError, l.ctx))}, nil
	}
	if contract.QtyStep.IsPositive() && !qty.Mod(contract.QtyStep).IsZero() {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.QuantityFormatError, i18n.Translate(i18n.QuantityFormatError, l.ctx))}, nil
	}
	if contract.PriceTick.IsPositive() && !price.Mod(contract.PriceTick).IsZero() {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.PriceFormatError, i18n.Translate(i18n.PriceFormatError, l.ctx))}, nil
	}

	marginAmount := decimal.Zero
	marginCoin := contract.SettleCoin
	// Every buy order pays premium, including buy-to-close. It must be funded
	// before entering matching; position effect only controls position changes.
	if in.Side == common.Side_SIDE_BUY {
		maxFeeRate := decimal.Max(contract.MakerFeeRate, contract.TakerFeeRate)
		marginAmount = optionTurnover(contract, price, qty).Mul(decimal.NewFromInt(1).Add(maxFeeRate)).Round(16)
		if in.OrderType == option.OrderType_ORDER_TYPE_MARKET {
			maxTurnover, parseErr := conv.ParseDecimalField(in.MaxTurnover)
			if parseErr != nil || !maxTurnover.IsPositive() || maxTurnover.LessThan(marginAmount) {
				return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
			}
			marginAmount = maxTurnover
		}
	} else if in.PositionEffect == option.PositionEffect_POSITION_EFFECT_OPEN {
		switch option.SellerMarginMode(contract.SellerMarginMode) {
		case option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED:
			market, findErr := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, tenantId, contract.Id)
			now := time.Now().Unix()
			if findErr != nil || !logichelpers.IsUnderlyingFresh(market, now, 30) {
				return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotTradable, i18n.Translate(i18n.ContractNotTradable, l.ctx))}, nil
			}
			marginAmount = optionSellerMargin(contract, market.UnderlyingPrice, price, qty, false)
			if !marginAmount.IsPositive() {
				return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
			}
		case option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO:
			// Account-level incremental risk is calculated under the risk-account
			// row lock in the order transaction below.
		case option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY:
			if contract.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL) {
				return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
			}
			if contract.OptionType == int64(option.OptionType_OPTION_TYPE_CALL) {
				marginCoin = contract.UnderlyingCoin
				marginAmount = qty.Mul(optionMultiplier(contract)).Round(16)
			} else {
				marginAmount = contract.StrikePrice.Mul(qty).Mul(optionMultiplier(contract)).Round(16)
			}
			if marginCoin == "" || !marginAmount.IsPositive() {
				return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
			}
		default:
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
		}
	}

	if in.ClientOrderId != "" {
		exists, err := l.svcCtx.OptionOrderModel.FindOneByTenantIdUserIdClientOrderId(l.ctx, tenantId, userId, in.ClientOrderId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if exists != nil {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ClientOrderIDAlreadyExists, i18n.Translate(i18n.ClientOrderIDAlreadyExists, l.ctx)), Data: &option.PlaceOrderData{OrderNo: exists.OrderNo, OrderId: exists.Id}}, nil
		}
	}

	orderNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "OP", "")
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	initialStatus := option.OrderStatus_ORDER_STATUS_PENDING
	if marginAmount.IsPositive() {
		initialStatus = option.OrderStatus_ORDER_STATUS_FUNDING
	}
	order := &models.TOptionOrder{
		TenantId:         tenantId,
		OrderNo:          orderNo,
		UserId:           userId,
		AccountId:        in.AccountId,
		ContractId:       in.ContractId,
		UnderlyingSymbol: contract.UnderlyingSymbol,
		Side:             int64(in.Side),
		PositionEffect:   int64(in.PositionEffect),
		OrderType:        int64(in.OrderType),
		Price:            price,
		Qty:              qty,
		FilledQty:        decimal.Zero,
		UnfilledQty:      qty,
		AvgPrice:         decimal.Zero,
		Turnover:         decimal.Zero,
		Fee:              decimal.Zero,
		FeeCoin:          contract.SettleCoin,
		MarginAmount:     marginAmount,
		MarginCoin:       marginCoin,
		Source:           int64(l.orderSource),
		ClientOrderId:    in.ClientOrderId,
		ReduceOnly:       int64(in.ReduceOnly),
		Mmp:              int64(common.YesNo_YES_NO_NO),
		MmpGroup:         mmpGroup,
		Status:           int64(initialStatus),
		CreateTimes:      now,
		UpdateTimes:      now,
	}
	if in.Mmp == common.YesNo_YES_NO_YES {
		order.Mmp = int64(common.YesNo_YES_NO_YES)
	}
	var id int64
	var controlRejection *orderControlRejection
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		clientOrderKeyModel := models.NewTOptionClientOrderKeyModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		lockedContract, rejection, err := evaluateOrderTradingControls(
			ctx, l.svcCtx, conn, order, now,
		)
		if err != nil {
			return err
		}
		if rejection != nil {
			controlRejection = rejection
			return nil
		}
		contract = lockedContract
		if contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) &&
			order.Side == int64(common.Side_SIDE_SELL) {
			portfolioResult, err := calculatePortfolioOrderMargin(
				ctx, l.svcCtx, conn, order, contract, now,
			)
			if err != nil {
				return err
			}
			marginAmount = portfolioResult.margin
			order.MarginAmount = portfolioResult.margin
			order.PortfolioRiskConfigId = portfolioResult.configID
			order.PortfolioRiskConfigVersion = portfolioResult.configVersion
			if portfolioResult.margin.IsPositive() {
				order.Status = int64(option.OrderStatus_ORDER_STATUS_FUNDING)
			} else {
				order.Status = int64(option.OrderStatus_ORDER_STATUS_PENDING)
			}
		}
		result, err := orderModel.Insert(ctx, order)
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		order.Id = id
		if order.ClientOrderId != "" {
			if _, err := clientOrderKeyModel.Insert(ctx, &models.TOptionClientOrderKey{
				TenantId: order.TenantId, UserId: order.UserId,
				ClientOrderId: order.ClientOrderId, OrderId: order.Id,
				OrderNo: order.OrderNo, CreateTimes: now,
			}); err != nil {
				return err
			}
		}
		if err := freezeClosePosition(ctx, positionModel, order, now); err != nil {
			return err
		}
		if marginAmount.IsPositive() {
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: tenantId, InstructionNo: order.OrderNo + "-FREEZE",
				BizNo: order.OrderNo, OrderId: order.Id, UserId: userId, AccountId: in.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
				TargetBizNo: order.OrderNo, Coin: marginCoin, Amount: marginAmount,
				StepNo: 1, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if in.ClientOrderId != "" {
			exists, findErr := l.svcCtx.OptionOrderModel.FindOneByTenantIdUserIdClientOrderId(l.ctx, tenantId, userId, in.ClientOrderId)
			if findErr == nil {
				return &option.PlaceOrderResp{
					Base: helper.ErrResp(i18n.ClientOrderIDAlreadyExists, i18n.Translate(i18n.ClientOrderIDAlreadyExists, l.ctx)),
					Data: &option.PlaceOrderData{OrderNo: exists.OrderNo, OrderId: exists.Id},
				}, nil
			}
			if !errors.Is(findErr, models.ErrNotFound) {
				return nil, findErr
			}
		}
		if errors.Is(err, models.ErrNotFound) {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.PositionNotFound, i18n.Translate(i18n.PositionNotFound, l.ctx)), Data: &option.PlaceOrderData{OrderNo: order.OrderNo, OrderId: id}}, nil
		}
		if i18n.IsStatusError(err, i18n.QuantityFormatError) {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.QuantityFormatError, i18n.Translate(i18n.QuantityFormatError, l.ctx)), Data: &option.PlaceOrderData{OrderNo: order.OrderNo, OrderId: id}}, nil
		}
		return nil, err
	}
	if controlRejection != nil {
		l.Infof(
			"option trading control metric event=%s reason=%s tenantId=%d userId=%d contractId=%d detail=%s",
			controlEventOrderRejected, controlRejection.reason, tenantId, userId, in.ContractId, controlRejection.detail,
		)
		return &option.PlaceOrderResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, controlRejection.reason),
		}, nil
	}

	if marginAmount.IsPositive() {
		publishOptionOrderChanged(l.ctx, l.svcCtx, order)
		return &option.PlaceOrderResp{Base: helper.OkResp(), Data: &option.PlaceOrderData{OrderNo: order.OrderNo, OrderId: id}}, nil
	}

	if err := l.matchOrder(contract, order); err != nil {
		return nil, err
	}
	publishOptionOrderChanged(l.ctx, l.svcCtx, order)

	return &option.PlaceOrderResp{Base: helper.OkResp(), Data: &option.PlaceOrderData{OrderNo: order.OrderNo, OrderId: id}}, nil
}
