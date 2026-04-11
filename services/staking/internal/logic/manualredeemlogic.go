package logic

import (
	"context"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualRedeemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualRedeemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualRedeemLogic {
	return &ManualRedeemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动赎回
func (l *ManualRedeemLogic) ManualRedeem(in *staking.AdminManualRedeemReq) (*staking.AdminManualRedeemResp, error) {
	order, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		return nil, err
	}
	if order == nil || order.TenantId != in.TenantId {
		return &staking.AdminManualRedeemResp{Page: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if order.Status == int64(staking.OrderStatus_ORDER_STATUS_REDEEMED) || order.Status == int64(staking.OrderStatus_ORDER_STATUS_EARLY_REDEEMED) || order.Status == int64(staking.OrderStatus_ORDER_STATUS_CANCELLED) {
		return &staking.AdminManualRedeemResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.StakingOrderCannotRedeem, l.ctx))}, nil
	}
	if in.RedeemType == staking.RedeemType_REDEEM_TYPE_EARLY && order.AllowEarlyRedeem != int64(staking.YesNo_YES_NO_YES) {
		return &staking.AdminManualRedeemResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.EarlyRedeemNotAllowed, l.ctx))}, nil
	}

	redeemAmount, err := conv.ParseFloatField(in.RedeemAmount)
	if err != nil || redeemAmount <= 0 || redeemAmount > order.StakeAmount {
		return &staking.AdminManualRedeemResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.RedeemAmountInvalid, l.ctx))}, nil
	}
	rewardAmount, err := conv.ParseFloatField(in.RewardAmount)
	if err != nil || rewardAmount < 0 {
		return &staking.AdminManualRedeemResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.RewardAmountInvalid, l.ctx))}, nil
	}
	feeRate, err := conv.ParseFloatField(in.FeeRate)
	if err != nil || feeRate < 0 {
		return &staking.AdminManualRedeemResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	feeAmount, err := conv.ParseFloatField(in.FeeAmount)
	if err != nil || feeAmount < 0 {
		return &staking.AdminManualRedeemResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	redeemNo := conv.GenerateBizNo("SKR")
	now := utils.NowMillis()
	if _, err := l.svcCtx.StakeRedeemLogModel.Insert(l.ctx, &models.TStakeRedeemLog{
		TenantId:     order.TenantId,
		OrderId:      order.Id,
		OrderNo:      order.OrderNo,
		Uid:          order.Uid,
		ProductId:    order.ProductId,
		RedeemNo:     redeemNo,
		RedeemType:   int64(in.RedeemType),
		StakeAmount:  order.StakeAmount,
		RedeemAmount: redeemAmount,
		RewardAmount: rewardAmount,
		FeeRate:      feeRate,
		FeeAmount:    feeAmount,
		RedeemStatus: int64(staking.RedeemStatus_REDEEM_STATUS_SUCCESS),
		RedeemTimes:  now,
		Remark:       in.Remark,
		CreateUserId: in.OperatorUid,
		UpdateUserId: in.OperatorUid,
		CreateTimes:  now,
		UpdateTimes:  now,
	}); err != nil {
		return nil, err
	}

	product, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, order.ProductId)
	if err == nil && product != nil {
		if product.StakedAmount >= order.StakeAmount {
			product.StakedAmount -= order.StakeAmount
		} else {
			product.StakedAmount = 0
		}
		product.UpdateUserId = in.OperatorUid
		product.UpdateTimes = now
		_ = l.svcCtx.StakeProductModel.Update(l.ctx, product)
	}

	order.RedeemAmount = redeemAmount
	order.RedeemFee = feeAmount
	order.TotalReward += rewardAmount
	order.PendingReward = 0
	order.RedeemType = int64(in.RedeemType)
	order.RedeemApplyTimes = now
	order.RedeemTimes = now
	if in.RedeemType == staking.RedeemType_REDEEM_TYPE_EARLY || in.RedeemType == staking.RedeemType_REDEEM_TYPE_MANUAL {
		order.Status = int64(staking.OrderStatus_ORDER_STATUS_EARLY_REDEEMED)
	} else {
		order.Status = int64(staking.OrderStatus_ORDER_STATUS_REDEEMED)
	}
	order.Remark = in.Remark
	order.UpdateUserId = in.OperatorUid
	order.UpdateTimes = now
	if err := l.svcCtx.StakeOrderModel.Update(l.ctx, order); err != nil {
		return nil, err
	}

	return &staking.AdminManualRedeemResp{Page: helper.OkResp(), Success: true, RedeemNo: redeemNo}, nil
}
