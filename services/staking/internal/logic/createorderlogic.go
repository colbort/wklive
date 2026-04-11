package logic

import (
	"context"
	"math"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
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
func (l *CreateOrderLogic) CreateOrder(in *staking.AppCreateOrderReq) (*staking.AppCreateOrderResp, error) {
	product, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, in.ProductId)
	if err != nil {
		return nil, err
	}
	if product == nil || product.TenantId != in.TenantId {
		return &staking.AppCreateOrderResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
	}
	if product.Status != int64(staking.ProductStatus_PRODUCT_STATUS_ENABLE) {
		return &staking.AppCreateOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.StakingProductUnavailable, l.ctx))}, nil
	}

	amount, err := conv.ParseFloatField(in.StakeAmount)
	if err != nil || amount <= 0 {
		return &staking.AppCreateOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.StakeAmountInvalid, l.ctx))}, nil
	}
	if product.MinAmount > 0 && amount < product.MinAmount {
		return &staking.AppCreateOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.StakeAmountBelowMinimum, l.ctx))}, nil
	}
	if product.MaxAmount > 0 && amount > product.MaxAmount {
		return &staking.AppCreateOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.StakeAmountAboveMaximum, l.ctx))}, nil
	}
	if product.StepAmount > 0 {
		steps := amount / product.StepAmount
		if math.Abs(steps-math.Round(steps)) > 1e-9 {
			return &staking.AppCreateOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.StakeAmountStepInvalid, l.ctx))}, nil
		}
	}
	if product.TotalAmount > 0 && product.StakedAmount+amount > product.TotalAmount+1e-9 {
		return &staking.AppCreateOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ProductQuotaInsufficient, l.ctx))}, nil
	}
	if product.UserLimitAmount > 0 {
		userStaked, err := l.svcCtx.StakeOrderModel.SumStakeAmountByStatuses(l.ctx, in.TenantId, in.Uid, in.ProductId, activeOrderStatuses())
		if err != nil {
			return nil, err
		}
		if userStaked+amount > product.UserLimitAmount+1e-9 {
			return &staking.AppCreateOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.UserStakeLimitExceeded, l.ctx))}, nil
		}
	}

	now := utils.NowMillis()
	endTimes := int64(0)
	if product.ProductType == int64(staking.ProductType_PRODUCT_TYPE_FIXED) && product.LockDays > 0 {
		endTimes = now + product.LockDays*24*3600*1000
	}
	orderNo := conv.GenerateBizNo("SKO")
	order := &models.TStakeOrder{
		TenantId:         in.TenantId,
		OrderNo:          orderNo,
		Uid:              in.Uid,
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
		NextRewardTimes:  calcNextRewardTime(int64(now), staking.RewardMode(product.RewardMode), int64(endTimes)),
		TotalReward:      0,
		PendingReward:    0,
		RedeemAmount:     0,
		RedeemFee:        0,
		Status:           int64(staking.OrderStatus_ORDER_STATUS_STAKING),
		RedeemType:       int64(staking.RedeemType_REDEEM_TYPE_NONE),
		RedeemApplyTimes: 0,
		RedeemTimes:      0,
		Source:           int64(in.Source),
		Remark:           in.Remark,
		CreateUserId:     in.Uid,
		UpdateUserId:     in.Uid,
		CreateTimes:      now,
		UpdateTimes:      now,
	}

	product.StakedAmount += amount
	product.UpdateUserId = in.Uid
	product.UpdateTimes = now
	if err := l.svcCtx.StakeProductModel.Update(l.ctx, product); err != nil {
		return nil, err
	}
	result, err := l.svcCtx.StakeOrderModel.Insert(l.ctx, order)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &staking.AppCreateOrderResp{Base: helper.OkResp(), Id: id, OrderNo: orderNo}, nil
}
