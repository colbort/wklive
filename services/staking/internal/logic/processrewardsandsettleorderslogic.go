package logic

import (
	"context"
	"errors"
	"fmt"

	"wklive/common/conv"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessRewardsAndSettleOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessRewardsAndSettleOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessRewardsAndSettleOrdersLogic {
	return &ProcessRewardsAndSettleOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 质押收益发放/到期结算
func (l *ProcessRewardsAndSettleOrdersLogic) ProcessRewardsAndSettleOrders(in *staking.StakingTaskReq) (*staking.StakingTaskResp, error) {
	return runStakingTaskWithLock(l.ctx, l.svcCtx, "process_rewards_and_settle_orders", func() (*staking.StakingTaskResp, error) {
		now := utils.NowMillis()
		cursor := int64(0)
		for {
			orders, _, err := l.svcCtx.StakeOrderModel.FindPage(l.ctx, in.GetTenantId(), cursor, 100, 0, 0, "", "", "", int64(staking.OrderStatus_ORDER_STATUS_STAKING), 0, 0, 0, 0, 0, 0)
			if err != nil {
				return nil, err
			}
			if len(orders) == 0 {
				break
			}
			for _, order := range orders {
				cursor = order.Id
				if err := l.processDailyReward(order, now); err != nil {
					return nil, err
				}
				if order.EndTimes > 0 && order.EndTimes <= now {
					if err := l.settleExpiredOrder(order, now); err != nil {
						return nil, err
					}
				}
			}
			if len(orders) < 100 {
				break
			}
		}
		return okStakingTaskResp(), nil
	})
}

func (l *ProcessRewardsAndSettleOrdersLogic) processDailyReward(order *models.TStakeOrder, now int64) error {
	if order.RewardMode != int64(staking.RewardMode_REWARD_MODE_DAILY) || order.NextRewardTimes == 0 || order.NextRewardTimes > now {
		return nil
	}
	rewardAmount := calcTaskReward(order, 1)
	if rewardAmount <= 0 {
		return nil
	}

	rewardNo := dailyRewardBizNo(order)
	resp, err := l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{
		TenantId:   order.TenantId,
		UserId:     order.UserId,
		WalletType: asset.WalletType_WALLET_TYPE_EARN,
		Coin:       order.RewardCoinSymbol,
		Amount:     conv.FloatString(rewardAmount),
		BizType:    asset.BizType_BIZ_TYPE_STAKING,
		SceneType:  asset.SceneType_SCENE_TYPE_STAKING_REWARD,
		BizId:      order.Id,
		BizNo:      rewardNo,
		Remark:     "staking daily reward task",
	})
	rewardStatus := int64(staking.RewardStatus_REWARD_STATUS_SUCCESS)
	remark := "staking daily reward task"
	if err != nil {
		l.Errorf("staking daily reward add asset rpc failed, tenantId=%d userId=%d orderId=%d orderNo=%s rewardNo=%s coin=%s amount=%v err=%v",
			order.TenantId, order.UserId, order.Id, order.OrderNo, rewardNo, order.RewardCoinSymbol, rewardAmount, err)
		rewardStatus = int64(staking.RewardStatus_REWARD_STATUS_FAIL)
		remark = err.Error()
	} else if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		l.Errorf("staking daily reward add asset failed, tenantId=%d userId=%d orderId=%d orderNo=%s rewardNo=%s coin=%s amount=%v msg=%s",
			order.TenantId, order.UserId, order.Id, order.OrderNo, rewardNo, order.RewardCoinSymbol, rewardAmount, assetBaseMsg(resp))
		rewardStatus = int64(staking.RewardStatus_REWARD_STATUS_FAIL)
		if resp != nil && resp.Base != nil {
			remark = resp.Base.Msg
		}
	}

	beforeReward := order.TotalReward
	if rewardStatus == int64(staking.RewardStatus_REWARD_STATUS_SUCCESS) {
		order.TotalReward += rewardAmount
		order.LastRewardTimes = now
		order.NextRewardTimes = calcNextRewardTime(now, staking.RewardMode(order.RewardMode), order.EndTimes)
		order.InterestDays++
	}
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
			AfterReward:      order.TotalReward,
			RewardType:       int64(staking.RewardType_REWARD_TYPE_DAILY),
			RewardStatus:     rewardStatus,
			RewardTimes:      now,
			Remark:           remark,
			CreateTimes:      now,
			UpdateTimes:      now,
		}); err != nil {
			return err
		}
		if rewardStatus == int64(staking.RewardStatus_REWARD_STATUS_SUCCESS) {
			return orderModel.Update(ctx, order)
		}
		return nil
	})
	if err != nil {
		l.Errorf("staking daily reward transaction failed, tenantId=%d userId=%d orderId=%d orderNo=%s rewardNo=%s coin=%s amount=%v status=%d err=%v",
			order.TenantId, order.UserId, order.Id, order.OrderNo, rewardNo, order.RewardCoinSymbol, rewardAmount, rewardStatus, err)
		return err
	}
	return nil
}

func (l *ProcessRewardsAndSettleOrdersLogic) settleExpiredOrder(order *models.TStakeOrder, now int64) error {
	if order.RewardMode == int64(staking.RewardMode_REWARD_MODE_MATURITY) {
		days := order.LockDays
		if days <= 0 {
			days = 1
		}
		order.PendingReward += calcTaskReward(order, days)
	}

	redeemNo := maturityRedeemBizNo(order)
	rewardAmount := order.PendingReward
	resp, err := l.svcCtx.AssetClient.UnlockAssetByBizNo(l.ctx, &asset.UnlockAssetByBizNoReq{
		TenantId:      order.TenantId,
		TargetBizType: asset.BizType_BIZ_TYPE_STAKING,
		TargetBizNo:   order.OrderNo,
		Amount:        conv.FloatString(order.StakeAmount),
		BizType:       asset.BizType_BIZ_TYPE_STAKING,
		SceneType:     asset.SceneType_SCENE_TYPE_STAKING_RELEASE,
		BizId:         order.Id,
		BizNo:         redeemNo,
		Remark:        "staking maturity redeem task",
	})
	if err != nil {
		l.Errorf("staking maturity redeem unlock asset rpc failed, tenantId=%d userId=%d orderId=%d orderNo=%s redeemNo=%s amount=%v err=%v",
			order.TenantId, order.UserId, order.Id, order.OrderNo, redeemNo, order.StakeAmount, err)
		return err
	}
	if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		l.Errorf("staking maturity redeem unlock asset failed, tenantId=%d userId=%d orderId=%d orderNo=%s redeemNo=%s amount=%v msg=%s",
			order.TenantId, order.UserId, order.Id, order.OrderNo, redeemNo, order.StakeAmount, assetBaseMsg(resp))
		return l.insertRedeemFailedLog(order, redeemNo, rewardAmount, now, assetBaseMsg(resp))
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
			Remark:     "staking maturity reward task",
		})
		if err != nil {
			l.Errorf("staking maturity reward add asset rpc failed, tenantId=%d userId=%d orderId=%d orderNo=%s redeemNo=%s coin=%s amount=%v err=%v",
				order.TenantId, order.UserId, order.Id, order.OrderNo, redeemNo, order.RewardCoinSymbol, rewardAmount, err)
			return err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
			l.Errorf("staking maturity reward add asset failed, tenantId=%d userId=%d orderId=%d orderNo=%s redeemNo=%s coin=%s amount=%v msg=%s",
				order.TenantId, order.UserId, order.Id, order.OrderNo, redeemNo, order.RewardCoinSymbol, rewardAmount, assetBaseMsg(resp))
			return l.insertRedeemFailedLog(order, redeemNo, rewardAmount, now, assetBaseMsg(resp))
		}
	}

	product, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, order.ProductId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if product != nil {
		if product.StakedAmount >= order.StakeAmount {
			product.StakedAmount -= order.StakeAmount
		} else {
			product.StakedAmount = 0
		}
		product.UpdateTimes = now
	}
	order.RedeemAmount = order.StakeAmount
	order.TotalReward += rewardAmount
	order.PendingReward = 0
	order.RedeemType = int64(staking.RedeemType_REDEEM_TYPE_MATURITY)
	order.RedeemApplyTimes = now
	order.RedeemTimes = now
	order.Status = int64(staking.OrderStatus_ORDER_STATUS_REDEEMED)
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
			RedeemType:   int64(staking.RedeemType_REDEEM_TYPE_MATURITY),
			StakeAmount:  order.StakeAmount,
			RedeemAmount: order.StakeAmount,
			RewardAmount: rewardAmount,
			RedeemStatus: int64(staking.RedeemStatus_REDEEM_STATUS_SUCCESS),
			RedeemTimes:  now,
			Remark:       "staking maturity redeem task",
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
		l.Errorf("staking maturity redeem transaction failed after asset change, tenantId=%d userId=%d orderId=%d orderNo=%s redeemNo=%s rewardAmount=%v err=%v",
			order.TenantId, order.UserId, order.Id, order.OrderNo, redeemNo, rewardAmount, err)
		return err
	}
	return nil
}

func (l *ProcessRewardsAndSettleOrdersLogic) insertRedeemFailedLog(order *models.TStakeOrder, redeemNo string, rewardAmount float64, now int64, remark string) error {
	if remark == "" {
		remark = "staking maturity redeem failed"
	}
	_, err := l.svcCtx.StakeRedeemLogModel.Insert(l.ctx, &models.TStakeRedeemLog{
		TenantId:     order.TenantId,
		OrderId:      order.Id,
		OrderNo:      order.OrderNo,
		UserId:       order.UserId,
		ProductId:    order.ProductId,
		RedeemNo:     redeemNo,
		RedeemType:   int64(staking.RedeemType_REDEEM_TYPE_MATURITY),
		StakeAmount:  order.StakeAmount,
		RedeemAmount: order.StakeAmount,
		RewardAmount: rewardAmount,
		RedeemStatus: int64(staking.RedeemStatus_REDEEM_STATUS_FAIL),
		RedeemTimes:  now,
		Remark:       remark,
		CreateTimes:  now,
		UpdateTimes:  now,
	})
	if err != nil {
		l.Errorf("staking redeem failed log insert failed, tenantId=%d userId=%d orderId=%d orderNo=%s redeemNo=%s rewardAmount=%v remark=%s err=%v",
			order.TenantId, order.UserId, order.Id, order.OrderNo, redeemNo, rewardAmount, remark, err)
	}
	return err
}

func assetBaseMsg(resp interface{ GetBase() *common.RespBase }) string {
	if resp == nil || resp.GetBase() == nil {
		return ""
	}
	return resp.GetBase().Msg
}

func dailyRewardBizNo(order *models.TStakeOrder) string {
	return fmt.Sprintf("SKW_%d_%d", order.Id, order.NextRewardTimes)
}

func maturityRedeemBizNo(order *models.TStakeOrder) string {
	return fmt.Sprintf("SKR_%d_%d", order.Id, order.EndTimes)
}
