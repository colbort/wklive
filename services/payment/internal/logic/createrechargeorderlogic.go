package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/notify"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	clientIP, _ := utils.GetClientIPFromMd(l.ctx)
	now := utils.NowMillis()

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
			Base: helper.GetErrResp(i18n.PaymentChannelNotFound, i18n.Translate(i18n.PaymentChannelNotFound, l.ctx)),
		}, nil
	}

	// 验证通道可用性
	if channel.Status != 1 {
		return &payment.CreateRechargeOrderResp{
			Base: helper.GetErrResp(i18n.PaymentChannelUnavailable, i18n.Translate(i18n.PaymentChannelUnavailable, l.ctx)),
		}, nil
	}
	platform, err := l.svcCtx.PayPlatformModel.FindOne(l.ctx, channel.PlatformId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	rechargeType := rechargeTypeFromPlatform(platform)

	// 验证金额限制
	if in.RechargeAmount < channel.SingleMinAmount || in.RechargeAmount > channel.SingleMaxAmount {
		return &payment.CreateRechargeOrderResp{
			Base: helper.GetErrResp(i18n.RechargeAmountOutOfLimit, i18n.Translate(i18n.RechargeAmountOutOfLimit, l.ctx)),
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
		TenantId:     tenantId,
		UserId:       userId,
		OrderNo:      orderNo,
		BizOrderNo:   sql.NullString{String: in.BizOrderNo, Valid: in.BizOrderNo != ""},
		PlatformId:   channel.PlatformId,
		ProductId:    channel.ProductId,
		AccountId:    channel.AccountId,
		ChannelId:    in.ChannelId,
		RechargeType: int64(rechargeType),
		Currency:     in.Currency,
		OrderAmount:  in.RechargeAmount,
		PayAmount:    in.RechargeAmount,
		FeeAmount:    feeAmount,
		Subject:      sql.NullString{String: in.Subject, Valid: in.Subject != ""},
		Body:         sql.NullString{String: in.Body, Valid: in.Body != ""},
		ClientType:   int64(in.ClientType),
		ClientIp:     sql.NullString{String: clientIP, Valid: clientIP != ""},
		Status:       int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING),
		CreateTimes:  now,
		UpdateTimes:  now,
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		rechargeOrderModel := models.NewTRechargeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.RechargeOrderModel)
		userRechargeStatModel := models.NewTUserRechargeStatModel(conn, l.svcCtx.Config.CacheRedis).(models.UserRechargeStatModel)

		if _, err := rechargeOrderModel.Insert(ctx, rechargeOrder); err != nil {
			return err
		}

		stat, err := userRechargeStatModel.FindOneByTenantIdUserId(ctx, tenantId, userId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if stat == nil {
			_, err = userRechargeStatModel.Insert(ctx, &models.TUserRechargeStat{
				TenantId:    tenantId,
				UserId:      userId,
				CreateTimes: now,
				UpdateTimes: now,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("Create recharge order success: %s, user_id: %d", orderNo, userId)
	event := notify.NewEvent(notify.EventTypeRecharge, notify.EventLevelInfo, "用户充值", fmt.Sprintf("用户 %d 发起充值订单 %s", userId, orderNo))
	event.Source = "payment"
	event.TenantID = tenantId
	event.UserID = userId
	event.BizNo = orderNo
	event.Data = map[string]any{
		"amount":   in.RechargeAmount,
		"currency": in.Currency,
	}
	if err := notify.Publish(l.ctx, l.svcCtx.Redis, event); err != nil {
		l.Errorf("publish admin recharge notification failed, orderNo=%s err=%v", orderNo, err)
	}

	return &payment.CreateRechargeOrderResp{
		Base: helper.OkResp(),
		Data: toRechargeOrderProto(rechargeOrder),
	}, nil
}
