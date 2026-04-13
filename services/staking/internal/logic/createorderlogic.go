package logic

import (
	"context"
	"math"

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

	orderNo, err := l.svcCtx.GenerateBizNo(l.ctx, "SKO")
	if err != nil {
		return nil, err
	}
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

	lockResp, err := l.svcCtx.AssetClient.LockAsset(l.ctx, &asset.LockAssetReq{
		TenantId:   in.TenantId,
		UserId:     in.Uid,
		WalletType: asset.WalletType_WALLET_TYPE_EARN,
		Coin:       product.CoinSymbol,
		Amount:     conv.FloatString(amount),
		BizType:    asset.BizType_BIZ_TYPE_STAKING,
		SceneType:  asset.SceneType_SCENE_TYPE_STAKING_JOIN,
		BizNo:      orderNo,
		StartTime:  now,
		EndTime:    endTimes,
		Remark:     in.Remark,
	})
	if err != nil {
		return nil, err
	}
	if lockResp == nil || lockResp.Base == nil || lockResp.Base.Code != 0 {
		if lockResp != nil && lockResp.Base != nil {
			return &staking.AppCreateOrderResp{Base: lockResp.Base}, nil
		}
		return nil, err
	}

	product.StakedAmount += amount
	product.UpdateUserId = in.Uid
	product.UpdateTimes = now
	var id int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		productModel := models.NewTStakeProductModel(conn, l.svcCtx.Config.CacheRedis).(models.StakeProductModel)
		orderModel := models.NewTStakeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.StakeOrderModel)

		if err := productModel.Update(ctx, product); err != nil {
			return err
		}
		result, err := orderModel.Insert(ctx, order)
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		return err
	})
	if err != nil {
		_, unlockErr := l.svcCtx.AssetClient.UnlockAssetByBizNo(l.ctx, &asset.UnlockAssetByBizNoReq{
			TenantId:      in.TenantId,
			TargetBizType: asset.BizType_BIZ_TYPE_STAKING,
			TargetBizNo:   orderNo,
			Amount:        conv.FloatString(amount),
			BizType:       asset.BizType_BIZ_TYPE_STAKING,
			SceneType:     asset.SceneType_SCENE_TYPE_STAKING_RELEASE,
			BizNo:         orderNo + "_rollback",
			Remark:        "staking create order rollback",
		})
		if unlockErr != nil {
			l.Errorf("rollback staking lock asset failed, orderNo=%s err=%v", orderNo, unlockErr)
		}
		return nil, err
	}

	return &staking.AppCreateOrderResp{Base: helper.OkResp(), Id: id, OrderNo: orderNo}, nil
}
