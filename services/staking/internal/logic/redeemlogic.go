package logic

import (
	"context"
	"errors"
	"fmt"

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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	order, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		return nil, err
	}
	if order == nil || order.TenantId != tenantId {
		return &staking.AppRedeemResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if order.UserId != userId {
		return &staking.AppRedeemResp{Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionAccessOrder, l.ctx))}, nil
	}
	if order.Status == int64(staking.OrderStatus_ORDER_STATUS_REDEEMED) || order.Status == int64(staking.OrderStatus_ORDER_STATUS_EARLY_REDEEMED) || order.Status == int64(staking.OrderStatus_ORDER_STATUS_CANCELLED) {
		return &staking.AppRedeemResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.StakingOrderCannotRedeem, l.ctx))}, nil
	}

	redeemType := in.RedeemType
	if redeemType == staking.RedeemType_REDEEM_TYPE_UNKNOWN {
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
	redeemNo, err := l.svcCtx.GenerateBizNo(l.ctx, "SKR")
	if err != nil {
		return nil, err
	}
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
			l.Errorf("staking redeem unlock asset rpc failed, tenantId=%d userId=%d orderNo=%s redeemNo=%s amount=%v err=%v",
				order.TenantId, order.UserId, order.OrderNo, redeemNo, unlockAmount, err)
			return nil, err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
			l.Errorf("staking redeem unlock asset failed, tenantId=%d userId=%d orderNo=%s redeemNo=%s amount=%v msg=%s",
				order.TenantId, order.UserId, order.OrderNo, redeemNo, unlockAmount, assetBaseMsg(resp))
			if resp != nil && resp.Base != nil {
				return &staking.AppRedeemResp{Base: resp.Base}, nil
			}
			return nil, fmt.Errorf("staking redeem unlock asset returned empty response")
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
			l.Errorf("staking redeem deduct locked fee rpc failed, tenantId=%d userId=%d orderNo=%s redeemNo=%s amount=%v err=%v",
				order.TenantId, order.UserId, order.OrderNo, redeemNo, feeAmount, err)
			return nil, err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
			l.Errorf("staking redeem deduct locked fee failed, tenantId=%d userId=%d orderNo=%s redeemNo=%s amount=%v msg=%s",
				order.TenantId, order.UserId, order.OrderNo, redeemNo, feeAmount, assetBaseMsg(resp))
			if resp != nil && resp.Base != nil {
				return &staking.AppRedeemResp{Base: resp.Base}, nil
			}
			return nil, fmt.Errorf("staking redeem deduct locked fee returned empty response")
		}
	}
	if rewardAmount > 0 {
		resp, err := l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{
			TenantId:   order.TenantId,
			UserId:     order.UserId,
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
			l.Errorf("staking redeem add reward rpc failed, tenantId=%d userId=%d orderNo=%s redeemNo=%s coin=%s amount=%v err=%v",
				order.TenantId, order.UserId, order.OrderNo, redeemNo, order.RewardCoinSymbol, rewardAmount, err)
			return nil, err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
			l.Errorf("staking redeem add reward failed, tenantId=%d userId=%d orderNo=%s redeemNo=%s coin=%s amount=%v msg=%s",
				order.TenantId, order.UserId, order.OrderNo, redeemNo, order.RewardCoinSymbol, rewardAmount, assetBaseMsg(resp))
			if resp != nil && resp.Base != nil {
				return &staking.AppRedeemResp{Base: resp.Base}, nil
			}
			return nil, fmt.Errorf("staking redeem add reward returned empty response")
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
		product.UpdateUserId = userId
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
	order.UpdateUserId = userId
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
			UserId:       order.UserId,
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
			CreateUserId: userId,
			UpdateUserId: userId,
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

	return &staking.AppRedeemResp{Base: helper.OkResp(), Success: 1, RedeemNo: redeemNo}, nil
}
