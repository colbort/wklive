package applogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建质押订单
func (l *CreateOrderLogic) CreateOrder(in *staking.CreateOrderReq) (*staking.CreateOrderResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	requestNo := strings.TrimSpace(in.RequestNo)
	if requestNo == "" || len(requestNo) > 96 {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	amount, err := conv.ParseDecimalField(in.StakeAmount)
	if err != nil || !amount.IsPositive() {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.StakeAmountInvalid, i18n.Translate(i18n.StakeAmountInvalid, l.ctx))}, nil
	}
	if existing, findErr := l.svcCtx.StakeOrderModel.FindOneByTenantIdUserIdRequestNo(l.ctx, tenantId, userId, requestNo); findErr == nil {
		if existing.ProductId != in.ProductId || !existing.StakeAmount.Equal(amount) || existing.Source != int64(in.Source) {
			return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		return createOrderExistingResp(l.ctx, existing), nil
	} else if !errors.Is(findErr, models.ErrNotFound) {
		return nil, findErr
	}

	product, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, in.ProductId)
	if err != nil {
		return nil, err
	}
	if product == nil || product.TenantId != tenantId {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.ProductNotFound, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
	}
	if product.Status != int64(staking.ProductStatus_PRODUCT_STATUS_ENABLE) {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.StakingProductUnavailable, i18n.Translate(i18n.StakingProductUnavailable, l.ctx))}, nil
	}
	if err := helpers.ValidateStakeProduct(product); err != nil {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.StakingProductUnavailable, i18n.Translate(i18n.StakingProductUnavailable, l.ctx))}, nil
	}
	if in.Source != staking.SourceType_SOURCE_TYPE_H5 && in.Source != staking.SourceType_SOURCE_TYPE_APP && in.Source != staking.SourceType_SOURCE_TYPE_API {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	if product.MinAmount.IsPositive() && amount.LessThan(product.MinAmount) {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.StakeAmountBelowMinimum, i18n.Translate(i18n.StakeAmountBelowMinimum, l.ctx))}, nil
	}
	if product.MaxAmount.IsPositive() && amount.GreaterThan(product.MaxAmount) {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.StakeAmountAboveMaximum, i18n.Translate(i18n.StakeAmountAboveMaximum, l.ctx))}, nil
	}
	if product.StepAmount.IsPositive() {
		if !amount.Mod(product.StepAmount).IsZero() {
			return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.StakeAmountStepInvalid, i18n.Translate(i18n.StakeAmountStepInvalid, l.ctx))}, nil
		}
	}
	now := utils.NowMillis()
	endTimes := int64(0)
	if product.ProductType == int64(staking.ProductType_PRODUCT_TYPE_FIXED) && product.LockDays > 0 {
		endTimes = now + product.LockDays*24*3600*1000
	}

	orderNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "SKO", "")
	if err != nil {
		return nil, err
	}
	order := &models.TStakeOrder{
		TenantId:         tenantId,
		OrderNo:          orderNo,
		RequestNo:        requestNo,
		UserId:           userId,
		ProductId:        product.Id,
		ProductNo:        product.ProductNo,
		ProductName:      product.ProductName,
		ProductType:      product.ProductType,
		CoinName:         product.CoinName,
		CoinSymbol:       product.CoinSymbol,
		RewardCoinName:   product.RewardCoinName,
		RewardCoinSymbol: product.RewardCoinSymbol,
		StakeAmount:      amount,
		Apr:              product.Apr,
		LockDays:         product.LockDays,
		InterestMode:     product.InterestMode,
		RewardMode:       product.RewardMode,
		AllowEarlyRedeem: product.AllowEarlyRedeem,
		EarlyRedeemRate:  product.EarlyRedeemRate,
		InterestDays:     0,
		StartTimes:       now,
		EndTimes:         endTimes,
		LastRewardTimes:  0,
		NextRewardTimes:  helpers.CalcNextRewardTime(int64(now), staking.RewardMode(product.RewardMode), int64(endTimes)),
		TotalReward:      decimal.Zero,
		PendingReward:    decimal.Zero,
		RedeemAmount:     decimal.Zero,
		RedeemFee:        decimal.Zero,
		Status:           int64(staking.OrderStatus_ORDER_STATUS_PENDING),
		RedeemType:       int64(staking.RedeemType_REDEEM_TYPE_NONE),
		RedeemApplyTimes: 0,
		RedeemTimes:      0,
		Source:           int64(in.Source),
		Remark:           in.Remark,
		CreateUserId:     userId,
		UpdateUserId:     userId,
		CreateTimes:      now,
		UpdateTimes:      now,
		Version:          1,
	}
	operation := &models.TStakeOperation{
		TenantId:        tenantId,
		UserId:          userId,
		OrderNo:         orderNo,
		OperationNo:     "SUBSCRIBE:" + orderNo,
		RequestNo:       requestNo,
		OperationType:   helpers.StakeOperationTypeSubscribe,
		PrincipalAmount: amount,
		RewardAmount:    decimal.Zero,
		FeeAmount:       decimal.Zero,
		PrincipalStatus: helpers.StakeOperationStepPending,
		RewardStatus:    helpers.StakeOperationStepNotRequired,
		FeeStatus:       helpers.StakeOperationStepNotRequired,
		Status:          helpers.StakeOperationStatusPending,
		OperatorUserId:  userId,
		Remark:          in.Remark,
		Version:         1,
		CreateTimes:     now,
		UpdateTimes:     now,
	}

	var id int64
	errQuota := errors.New("staking product quota insufficient")
	errUserLimit := errors.New("staking user product limit exceeded")
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		productModel := models.NewTStakeProductModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTStakeUserPositionModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTStakeOrderModel(conn, l.svcCtx.Config.CacheRedis)
		operationModel := models.NewTStakeOperationModel(conn, l.svcCtx.Config.CacheRedis)

		reserved, err := productModel.ReserveStakeAmount(ctx, product.Id, amount, now)
		if err != nil {
			return err
		}
		if !reserved {
			return errQuota
		}
		if err := positionModel.Ensure(ctx, tenantId, userId, product.Id, now); err != nil {
			return err
		}
		reserved, err = positionModel.ReserveAmount(ctx, tenantId, userId, product.Id, amount, product.UserLimitAmount, now)
		if err != nil {
			return err
		}
		if !reserved {
			return errUserLimit
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
		operation.OrderId = id
		_, err = operationModel.Insert(ctx, operation)
		return err
	})
	if errors.Is(err, errQuota) {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.ProductQuotaInsufficient, i18n.Translate(i18n.ProductQuotaInsufficient, l.ctx))}, nil
	}
	if errors.Is(err, errUserLimit) {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.UserStakeLimitExceeded, i18n.Translate(i18n.UserStakeLimitExceeded, l.ctx))}, nil
	}
	if err != nil {
		if existing, findErr := l.svcCtx.StakeOrderModel.FindOneByTenantIdUserIdRequestNo(l.ctx, tenantId, userId, requestNo); findErr == nil {
			return createOrderExistingResp(l.ctx, existing), nil
		}
		return nil, err
	}
	operation, err = l.svcCtx.StakeOperationModel.FindOneByTenantIdOperationNo(l.ctx, tenantId, operation.OperationNo)
	if err != nil {
		return nil, err
	}
	if err := helpers.ExecuteSubscribeOperation(l.ctx, l.svcCtx, operation); err != nil {
		if errors.Is(err, helpers.ErrStakeOperationProcessing) {
			return createOrderExistingResp(l.ctx, order), nil
		}
		return nil, err
	}
	order, err = l.svcCtx.StakeOrderModel.FindOne(l.ctx, id)
	if err != nil {
		return nil, err
	}
	publishStakingOrderChanged(l.ctx, l.svcCtx, order)

	return &staking.CreateOrderResp{
		Base: helper.OkResp(),
		Data: &staking.CreateOrderData{
			Id:      id,
			OrderNo: orderNo,
		},
	}, nil
}

func createOrderExistingResp(ctx context.Context, order *models.TStakeOrder) *staking.CreateOrderResp {
	base := helper.OkResp()
	if order.Status == int64(staking.OrderStatus_ORDER_STATUS_PENDING) {
		base = helper.ErrResp(i18n.AssetRequestProcessing, i18n.Translate(i18n.AssetRequestProcessing, ctx))
	}
	return &staking.CreateOrderResp{Base: base, Data: &staking.CreateOrderData{Id: order.Id, OrderNo: order.OrderNo}}
}
