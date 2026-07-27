package applogic

import (
	"context"
	"time"

	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/staking"
	"wklive/services/staking/internal/delayqueue"
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
	product, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, in.ProductId)
	if err != nil {
		return nil, err
	}
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	if product == nil || product.TenantId != tenantId {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.ProductNotFound, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
	}
	if product.Status != int64(staking.ProductStatus_PRODUCT_STATUS_ENABLE) {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.StakingProductUnavailable, i18n.Translate(i18n.StakingProductUnavailable, l.ctx))}, nil
	}

	amount, err := conv.ParseDecimalField(in.StakeAmount)
	if err != nil || !amount.IsPositive() {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.StakeAmountInvalid, i18n.Translate(i18n.StakeAmountInvalid, l.ctx))}, nil
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
	if product.TotalAmount.IsPositive() && product.StakedAmount.Add(amount).GreaterThan(product.TotalAmount) {
		return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.ProductQuotaInsufficient, i18n.Translate(i18n.ProductQuotaInsufficient, l.ctx))}, nil
	}
	if product.UserLimitAmount.IsPositive() {
		userStaked, err := l.svcCtx.StakeOrderModel.SumStakeAmountByStatuses(l.ctx, tenantId, userId, in.ProductId, activeOrderStatuses())
		if err != nil {
			return nil, err
		}
		if userStaked.Add(amount).GreaterThan(product.UserLimitAmount) {
			return &staking.CreateOrderResp{Base: helper.ErrResp(i18n.UserStakeLimitExceeded, i18n.Translate(i18n.UserStakeLimitExceeded, l.ctx))}, nil
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
		NextRewardTimes:  calcNextRewardTime(int64(now), staking.RewardMode(product.RewardMode), int64(endTimes)),
		TotalReward:      decimal.Zero,
		PendingReward:    decimal.Zero,
		RedeemAmount:     decimal.Zero,
		RedeemFee:        decimal.Zero,
		Status:           int64(staking.OrderStatus_ORDER_STATUS_STAKING),
		RedeemType:       int64(staking.RedeemType_REDEEM_TYPE_NONE),
		RedeemApplyTimes: 0,
		RedeemTimes:      0,
		Source:           int64(in.Source),
		Remark:           in.Remark,
		CreateUserId:     userId,
		UpdateUserId:     userId,
		CreateTimes:      now,
		UpdateTimes:      now,
	}

	lockResp, err := l.svcCtx.AssetClient.LockAsset(l.ctx, &asset.LockAssetReq{
		TenantId:   tenantId,
		UserId:     userId,
		WalletType: common.WalletType_WALLET_TYPE_EARN,
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
		l.Errorf("staking create order lock asset rpc failed, tenantId=%d userId=%d orderNo=%s coin=%s amount=%v err=%v",
			tenantId, userId, orderNo, product.CoinSymbol, amount, err)
		return nil, err
	}
	if lockResp == nil || lockResp.Base == nil || lockResp.Base.Code != 200 {
		l.Errorf("staking create order lock asset failed, tenantId=%d userId=%d orderNo=%s coin=%s amount=%v msg=%s",
			tenantId, userId, orderNo, product.CoinSymbol, amount, assetBaseMsg(lockResp))
		if lockResp != nil && lockResp.Base != nil {
			return &staking.CreateOrderResp{Base: lockResp.Base}, nil
		}
		return nil, i18n.StatusError(l.ctx, i18n.InternalServerError)
	}

	product.StakedAmount = product.StakedAmount.Add(amount)
	product.UpdateUserId = userId
	product.UpdateTimes = now
	var id int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		productModel := models.NewTStakeProductModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTStakeOrderModel(conn, l.svcCtx.Config.CacheRedis)

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
		l.Errorf("staking create order transaction failed after lock asset, tenantId=%d userId=%d orderNo=%s amount=%v err=%v",
			tenantId, userId, orderNo, amount, err)
		_, unlockErr := l.svcCtx.AssetClient.UnlockAssetByBizNo(l.ctx, &asset.UnlockAssetByBizNoReq{
			TenantId:      tenantId,
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
	if order.EndTimes > 0 && l.svcCtx.DelayQueue != nil {
		if enqueueErr := l.svcCtx.DelayQueue.At(delayqueue.Message{
			Action: delayqueue.ActionMatureOrder, TenantID: order.TenantId,
			OrderID: id, DueAt: order.EndTimes,
		}, time.UnixMilli(order.EndTimes)); enqueueErr != nil {
			l.Errorf("enqueue staking maturity failed, orderId=%d err=%v", id, enqueueErr)
		}
	}
	order.Id = id
	publishStakingOrderChanged(l.ctx, l.svcCtx, order)

	return &staking.CreateOrderResp{
		Base: helper.OkResp(),
		Data: &staking.CreateOrderData{
			Id:      id,
			OrderNo: orderNo,
		},
	}, nil
}
