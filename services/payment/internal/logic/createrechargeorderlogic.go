package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateRechargeOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRechargeOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRechargeOrderLogic {
	return &CreateRechargeOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建充值订单
func (l *CreateRechargeOrderLogic) CreateRechargeOrder(in *payment.CreateRechargeOrderReq) (*payment.CreateRechargeOrderResp, error) {
	now := time.Now().UnixMilli()

	// 生成订单号
	orderNo, err := l.svcCtx.GenerateOrderNo(l.ctx, "RC")
	if err != nil {
		return nil, err
	}

	// 查询通道信息
	channel, err := l.svcCtx.TenantPayChannelModel.FindOne(l.ctx, in.ChannelId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if channel == nil {
		return &payment.CreateRechargeOrderResp{
			Base: helper.GetErrResp(404, "支付通道不存在"),
		}, nil
	}

	// 验证通道可用性
	if channel.Status != 1 {
		return &payment.CreateRechargeOrderResp{
			Base: helper.GetErrResp(201, "支付通道暂不可用"),
		}, nil
	}

	// 验证金额限制
	if in.RechargeAmount < channel.SingleMinAmount || in.RechargeAmount > channel.SingleMaxAmount {
		return &payment.CreateRechargeOrderResp{
			Base: helper.GetErrResp(201, "充值金额超出限制"),
		}, nil
	}

	// 计算手续费
	var feeAmount int64
	if channel.FeeType == int64(payment.FeeType_FEE_TYPE_RATE) {
		// 按比例计算
		feeAmount = in.RechargeAmount * int64(channel.FeeRate*100) / 10000
	} else if channel.FeeType == int64(payment.FeeType_FEE_TYPE_FIXED) {
		// 固定费用
		feeAmount = channel.FeeFixedAmount
	}

	// 创建充值订单
	rechargeOrder := &models.TRechargeOrder{
		TenantId:    in.TenantId,
		UserId:      in.UserId,
		OrderNo:     orderNo,
		BizOrderNo:  sql.NullString{String: in.BizOrderNo, Valid: in.BizOrderNo != ""},
		PlatformId:  channel.PlatformId,
		ProductId:   channel.ProductId,
		AccountId:   channel.AccountId,
		ChannelId:   in.ChannelId,
		Currency:    in.Currency,
		OrderAmount: in.RechargeAmount,
		PayAmount:   in.RechargeAmount,
		FeeAmount:   feeAmount,
		Subject:     sql.NullString{String: in.Subject, Valid: in.Subject != ""},
		Body:        sql.NullString{String: in.Body, Valid: in.Body != ""},
		ClientType:  int64(in.ClientType),
		ClientIp:    sql.NullString{String: in.ClientIp, Valid: in.ClientIp != ""},
		Status:      int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING),
		CreateTimes: now,
		UpdateTimes: now,
	}

	_, err = l.svcCtx.RechargeOrderModel.Insert(l.ctx, rechargeOrder)
	if err != nil {
		return nil, err
	}

	// 更新用户充值统计
	l.updateUserRechargeStat(in.TenantId, in.UserId)

	l.Logger.Infof("Create recharge order success: %s, user_id: %d", orderNo, in.UserId)

	return &payment.CreateRechargeOrderResp{
		Base: helper.OkResp(),
		Order: &payment.RechargeOrder{
			Id:          rechargeOrder.Id,
			TenantId:    rechargeOrder.TenantId,
			UserId:      rechargeOrder.UserId,
			OrderNo:     rechargeOrder.OrderNo,
			BizOrderNo:  rechargeOrder.BizOrderNo.String,
			PlatformId:  rechargeOrder.PlatformId,
			ProductId:   rechargeOrder.ProductId,
			AccountId:   rechargeOrder.AccountId,
			ChannelId:   rechargeOrder.ChannelId,
			Currency:    rechargeOrder.Currency,
			OrderAmount: rechargeOrder.OrderAmount,
			PayAmount:   rechargeOrder.PayAmount,
			FeeAmount:   rechargeOrder.FeeAmount,
			Subject:     rechargeOrder.Subject.String,
			Body:        rechargeOrder.Body.String,
			ClientType:  payment.ClientType(rechargeOrder.ClientType),
			ClientIp:    rechargeOrder.ClientIp.String,
			Status:      payment.PayOrderStatus(rechargeOrder.Status),
			CreateTimes: rechargeOrder.CreateTimes,
			UpdateTimes: rechargeOrder.UpdateTimes,
		},
	}, nil
}

func (l *CreateRechargeOrderLogic) updateUserRechargeStat(tenantId, userId int64) {
	stat, err := l.svcCtx.UserRechargeStatModel.FindOneByTenantIdUserId(l.ctx, tenantId, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("Find user recharge stat error: %v", err)
		return
	}

	now := time.Now().UnixMilli()
	if stat == nil {
		// Create new stat
		newStat := &models.TUserRechargeStat{
			TenantId:    tenantId,
			UserId:      userId,
			CreateTimes: now,
			UpdateTimes: now,
		}
		_, err = l.svcCtx.UserRechargeStatModel.Insert(l.ctx, newStat)
		if err != nil {
			l.Logger.Errorf("Insert user recharge stat error: %v", err)
		}
	}
}
