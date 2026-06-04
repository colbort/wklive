package logic

import (
	"context"

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

type ManualRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualRewardLogic {
	return &ManualRewardLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动发放收益
func (l *ManualRewardLogic) ManualReward(in *staking.AdminManualRewardReq) (*staking.AdminManualRewardResp, error) {
	order, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		return nil, err
	}
	if order == nil || order.TenantId != in.TenantId {
		return &staking.AdminManualRewardResp{Page: helper.GetErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}

	rewardAmount, err := conv.ParseFloatField(in.RewardAmount)
	if err != nil || rewardAmount <= 0 {
		return &staking.AdminManualRewardResp{Page: helper.GetErrResp(i18n.RewardAmountInvalid, i18n.Translate(i18n.RewardAmountInvalid, l.ctx))}, nil
	}

	now := utils.NowMillis()
	resp, err := l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{
		TenantId:   order.TenantId,
		UserId:     order.UserId,
		WalletType: asset.WalletType_WALLET_TYPE_EARN,
		Coin:       order.RewardCoinSymbol,
		Amount:     conv.FloatString(rewardAmount),
		BizType:    asset.BizType_BIZ_TYPE_STAKING,
		SceneType:  asset.SceneType_SCENE_TYPE_STAKING_REWARD,
		BizId:      order.Id,
		BizNo:      order.OrderNo,
		Remark:     in.Remark,
	})
	if err != nil {
		l.Errorf("staking manual reward add asset rpc failed, tenantId=%d userId=%d orderId=%d orderNo=%s coin=%s amount=%v err=%v",
			order.TenantId, order.UserId, order.Id, order.OrderNo, order.RewardCoinSymbol, rewardAmount, err)
		return nil, err
	}
	if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		l.Errorf("staking manual reward add asset failed, tenantId=%d userId=%d orderId=%d orderNo=%s coin=%s amount=%v msg=%s",
			order.TenantId, order.UserId, order.Id, order.OrderNo, order.RewardCoinSymbol, rewardAmount, assetBaseMsg(resp))
		if resp != nil && resp.Base != nil {
			return &staking.AdminManualRewardResp{Page: resp.Base}, nil
		}
		return nil, i18n.StatusError(l.ctx, i18n.InternalServerError)
	}

	beforeReward := order.TotalReward
	afterReward := beforeReward + rewardAmount
	order.TotalReward = afterReward
	order.LastRewardTimes = now
	order.NextRewardTimes = calcNextRewardTime(int64(now), staking.RewardMode(order.RewardMode), int64(order.EndTimes))
	order.UpdateUserId = in.OperatorUid
	order.UpdateTimes = now
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		rewardLogModel := models.NewTStakeRewardLogModel(conn, l.svcCtx.Config.CacheRedis).(models.StakeRewardLogModel)
		orderModel := models.NewTStakeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.StakeOrderModel)

		if _, err := rewardLogModel.Insert(ctx, &models.TStakeRewardLog{
			TenantId:         order.TenantId,
			OrderId:          order.Id,
			OrderNo:          order.OrderNo,
			UserId:           order.UserId,
			ProductId:        order.ProductId,
			ProductName:      order.ProductName,
			CoinSymbol:       order.CoinSymbol,
			RewardCoinSymbol: order.RewardCoinSymbol,
			RewardAmount:     rewardAmount,
			BeforeReward:     beforeReward,
			AfterReward:      afterReward,
			RewardType:       int64(in.RewardType),
			RewardStatus:     int64(staking.RewardStatus_REWARD_STATUS_SUCCESS),
			RewardTimes:      now,
			Remark:           in.Remark,
			CreateUserId:     in.OperatorUid,
			UpdateUserId:     in.OperatorUid,
			CreateTimes:      now,
			UpdateTimes:      now,
		}); err != nil {
			return err
		}
		return orderModel.Update(ctx, order)
	})
	if err != nil {
		l.Errorf("staking manual reward transaction failed after add asset, tenantId=%d userId=%d orderId=%d orderNo=%s coin=%s amount=%v err=%v",
			order.TenantId, order.UserId, order.Id, order.OrderNo, order.RewardCoinSymbol, rewardAmount, err)
		return nil, err
	}

	return &staking.AdminManualRewardResp{Page: helper.OkResp(), Data: 1}, nil
}
