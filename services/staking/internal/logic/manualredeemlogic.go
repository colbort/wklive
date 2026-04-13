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
			return nil, err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 0 {
			if resp != nil && resp.Base != nil {
				return &staking.AdminManualRedeemResp{Page: resp.Base}, nil
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
			Remark:        "staking manual redeem fee deduct",
		})
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 0 {
			if resp != nil && resp.Base != nil {
				return &staking.AdminManualRedeemResp{Page: resp.Base}, nil
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
				return &staking.AdminManualRedeemResp{Page: resp.Base}, nil
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
		product.UpdateUserId = in.OperatorUid
		product.UpdateTimes = now
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

	return &staking.AdminManualRedeemResp{Page: helper.OkResp(), Success: 1, RedeemNo: redeemNo}, nil
}
