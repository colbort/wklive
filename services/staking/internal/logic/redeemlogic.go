package logic

import (
	"context"
	"errors"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RedeemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRedeemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RedeemLogic {
	return &RedeemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发起赎回
func (l *RedeemLogic) Redeem(in *staking.AppRedeemReq) (*staking.AppRedeemResp, error) {
	order, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		return nil, err
	}
	if order == nil || order.TenantId != in.TenantId {
		return &staking.AppRedeemResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if order.Uid != in.Uid {
		return &staking.AppRedeemResp{Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionAccessOrder, l.ctx))}, nil
	}
	if order.Status == int64(staking.OrderStatus_ORDER_STATUS_REDEEMED) || order.Status == int64(staking.OrderStatus_ORDER_STATUS_EARLY_REDEEMED) || order.Status == int64(staking.OrderStatus_ORDER_STATUS_CANCELLED) {
		return &staking.AppRedeemResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.StakingOrderCannotRedeem, l.ctx))}, nil
	}

	redeemType := in.RedeemType
	if redeemType == staking.RedeemType_REDEEM_TYPE_UNSPECIFIED {
		if order.Status == int64(staking.OrderStatus_ORDER_STATUS_EXPIRED) {
			redeemType = staking.RedeemType_REDEEM_TYPE_MATURITY
		} else {
			redeemType = staking.RedeemType_REDEEM_TYPE_EARLY
		}
	}
	if redeemType == staking.RedeemType_REDEEM_TYPE_EARLY && order.AllowEarlyRedeem != int64(staking.YesNo_YES_NO_YES) {
		return &staking.AppRedeemResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.EarlyRedeemNotAllowed, l.ctx))}, nil
	}

	redeemAmount := order.StakeAmount
	feeRate := 0.0
	if redeemType == staking.RedeemType_REDEEM_TYPE_EARLY {
		feeRate = order.EarlyRedeemRate
	}
	feeAmount := redeemAmount * feeRate / 100
	rewardAmount := order.PendingReward
	redeemNo := conv.GenerateBizNo("SKR")
	now := utils.NowMillis()
	unlockAmount := redeemAmount - feeAmount
	if unlockAmount < 0 {
		unlockAmount = 0
	}
	if unlockAmount > 0 {
		resp, err := l.svcCtx.AssetClient.UnlockAssetByBizNo(l.ctx, &asset.UnlockAssetByBizNoReq{
			TenantId:      order.TenantId,
			TargetBizType: asset.BizType_BIZ_TYPE_STAKING,
			TargetBizNo:   order.OrderNo,
			Amount:        conv.FloatString(unlockAmount),
			BizType:       asset.BizType_BIZ_TYPE_STAKING,
			SceneType:     asset.SceneType_SCENE_TYPE_STAKING_RELEASE,
			BizId:         order.Id,
			BizNo:         redeemNo,
			Remark:        in.Remark,
		})
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 0 {
			if resp != nil && resp.Base != nil {
				return &staking.AppRedeemResp{Base: resp.Base}, nil
			}
			return nil, err
		}
	}
	if feeAmount > 0 {
		resp, err := l.svcCtx.AssetClient.DeductLockedAssetByBizNo(l.ctx, &asset.DeductLockedAssetByBizNoReq{
			TenantId:      order.TenantId,
			TargetBizType: asset.BizType_BIZ_TYPE_STAKING,
			TargetBizNo:   order.OrderNo,
			Amount:        conv.FloatString(feeAmount),
			BizType:       asset.BizType_BIZ_TYPE_STAKING,
			SceneType:     asset.SceneType_SCENE_TYPE_STAKING_RELEASE,
			BizId:         order.Id,
			BizNo:         redeemNo,
			Remark:        "staking redeem fee deduct",
		})
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 0 {
			if resp != nil && resp.Base != nil {
				return &staking.AppRedeemResp{Base: resp.Base}, nil
			}
			return nil, err
		}
	}
	if rewardAmount > 0 {
		resp, err := l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{
			TenantId:   order.TenantId,
			UserId:     order.Uid,
			WalletType: asset.WalletType_WALLET_TYPE_EARN,
			Coin:       order.RewardCoinSymbol,
			Amount:     conv.FloatString(rewardAmount),
			BizType:    asset.BizType_BIZ_TYPE_STAKING,
			SceneType:  asset.SceneType_SCENE_TYPE_STAKING_REWARD,
			BizId:      order.Id,
			BizNo:      redeemNo,
			Remark:     in.Remark,
		})
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 0 {
			if resp != nil && resp.Base != nil {
				return &staking.AppRedeemResp{Base: resp.Base}, nil
			}
			return nil, err
		}
	}

	product, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, order.ProductId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if product != nil {
		if product.StakedAmount >= order.StakeAmount {
			product.StakedAmount -= order.StakeAmount
		} else {
			product.StakedAmount = 0
		}
		product.UpdateUserId = in.Uid
		product.UpdateTimes = now
	}

	order.RedeemAmount = redeemAmount
	order.RedeemFee = feeAmount
	order.TotalReward += rewardAmount
	order.PendingReward = 0
	order.RedeemType = int64(redeemType)
	order.RedeemApplyTimes = now
	order.RedeemTimes = now
	if redeemType == staking.RedeemType_REDEEM_TYPE_EARLY {
		order.Status = int64(staking.OrderStatus_ORDER_STATUS_EARLY_REDEEMED)
	} else {
		order.Status = int64(staking.OrderStatus_ORDER_STATUS_REDEEMED)
	}
	order.Remark = in.Remark
	order.UpdateUserId = in.Uid
	order.UpdateTimes = now
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		redeemLogModel := models.NewTStakeRedeemLogModel(conn, l.svcCtx.Config.CacheRedis).(models.StakeRedeemLogModel)
		productModel := models.NewTStakeProductModel(conn, l.svcCtx.Config.CacheRedis).(models.StakeProductModel)
		orderModel := models.NewTStakeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.StakeOrderModel)

		if _, err := redeemLogModel.Insert(ctx, &models.TStakeRedeemLog{
			TenantId:     order.TenantId,
			OrderId:      order.Id,
			OrderNo:      order.OrderNo,
			Uid:          order.Uid,
			ProductId:    order.ProductId,
			RedeemNo:     redeemNo,
			RedeemType:   int64(redeemType),
			StakeAmount:  order.StakeAmount,
			RedeemAmount: redeemAmount,
			RewardAmount: rewardAmount,
			FeeRate:      feeRate,
			FeeAmount:    feeAmount,
			RedeemStatus: int64(staking.RedeemStatus_REDEEM_STATUS_SUCCESS),
			RedeemTimes:  now,
			Remark:       in.Remark,
			CreateUserId: in.Uid,
			UpdateUserId: in.Uid,
			CreateTimes:  now,
			UpdateTimes:  now,
		}); err != nil {
			return err
		}
		if product != nil {
			if err := productModel.Update(ctx, product); err != nil {
				return err
			}
		}
		return orderModel.Update(ctx, order)
	})
	if err != nil {
		return nil, err
	}

	return &staking.AppRedeemResp{Base: helper.OkResp(), Success: true, RedeemNo: redeemNo}, nil
}
