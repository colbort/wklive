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
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type PlaceOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceOrderLogic {
	return &PlaceOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
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
	if contract.IsDeleted == int64(common.YesNo_YES_NO_YES) ||
		time.Now().Unix() < contract.ListTime ||
		(contract.ExpireTime > 0 && time.Now().Unix() >= contract.ExpireTime) {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotTradable, i18n.Translate(i18n.ContractNotTradable, l.ctx))}, nil
	}
	riskAccount, riskErr := l.svcCtx.OptionRiskAccountModel.FindOneByTenantIdUserIdAccountIdSettleCoin(
		l.ctx, tenantId, userId, in.AccountId, contract.SettleCoin,
	)
	if riskErr != nil && !errors.Is(riskErr, models.ErrNotFound) {
		return nil, riskErr
	}
	if riskErr == nil {
		switch option.RiskAccountStatus(riskAccount.Status) {
		case option.RiskAccountStatus_RISK_ACCOUNT_STATUS_LIQUIDATING,
			option.RiskAccountStatus_RISK_ACCOUNT_STATUS_BANKRUPT,
			option.RiskAccountStatus_RISK_ACCOUNT_STATUS_RESTRICTED:
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
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
		if contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
		}
		market, findErr := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, tenantId, contract.Id)
		now := time.Now().Unix()
		if findErr != nil || market.UnderlyingPrice.IsPositive() == false ||
			market.SnapshotTime <= 0 || market.SnapshotTime > now || now-market.SnapshotTime > 30 {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotTradable, i18n.Translate(i18n.ContractNotTradable, l.ctx))}, nil
		}
		marginAmount = optionSellerMargin(contract, market.UnderlyingPrice, price, qty, false)
		if !marginAmount.IsPositive() {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
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
		Source:           int64(option.OrderSource_ORDER_SOURCE_APP),
		ClientOrderId:    in.ClientOrderId,
		ReduceOnly:       int64(in.ReduceOnly),
		Mmp:              int64(in.Mmp),
		Status:           int64(initialStatus),
		CreateTimes:      now,
		UpdateTimes:      now,
	}
	var id int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		clientOrderKeyModel := models.NewTOptionClientOrderKeyModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
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
				TargetBizNo: order.OrderNo, Coin: contract.SettleCoin, Amount: marginAmount,
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
