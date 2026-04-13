package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
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
			Base: helper.GetErrResp(404, i18n.Translate(i18n.PaymentChannelNotFound, l.ctx)),
		}, nil
	}

	// 验证通道可用性
	if channel.Status != 1 {
		return &payment.CreateRechargeOrderResp{
			Base: helper.GetErrResp(201, i18n.Translate(i18n.PaymentChannelUnavailable, l.ctx)),
		}, nil
	}

	// 验证金额限制
	if in.RechargeAmount < channel.SingleMinAmount || in.RechargeAmount > channel.SingleMaxAmount {
		return &payment.CreateRechargeOrderResp{
			Base: helper.GetErrResp(201, i18n.Translate(i18n.RechargeAmountOutOfLimit, l.ctx)),
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

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		rechargeOrderModel := models.NewTRechargeOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.RechargeOrderModel)
		userRechargeStatModel := models.NewTUserRechargeStatModel(conn, l.svcCtx.Config.CacheRedis).(models.UserRechargeStatModel)

		if _, err := rechargeOrderModel.Insert(ctx, rechargeOrder); err != nil {
			return err
		}

		stat, err := userRechargeStatModel.FindOneByTenantIdUserId(ctx, in.TenantId, in.UserId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if stat == nil {
			_, err = userRechargeStatModel.Insert(ctx, &models.TUserRechargeStat{
				TenantId:    in.TenantId,
				UserId:      in.UserId,
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

	l.Logger.Infof("Create recharge order success: %s, user_id: %d", orderNo, in.UserId)

	return &payment.CreateRechargeOrderResp{
		Base: helper.OkResp(),
		Order: toRechargeOrderProto(rechargeOrder),
	}, nil
}
