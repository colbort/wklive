package logic

import (
	"context"
	"database/sql"
	"fmt"

	"wklive/common/helper"
	"wklive/common/notify"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateWithdrawOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateWithdrawOrderLogic {
	return &CreateWithdrawOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提现
func (l *CreateWithdrawOrderLogic) CreateWithdrawOrder(in *payment.CreateWithdrawOrderReq) (*payment.CreateWithdrawOrderResp, error) {
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
	orderNo, err := l.svcCtx.GenerateOrderNo(l.ctx, "WD")
	if err != nil {
		return nil, err
	}

	// Create withdraw order
	withdrawOrder := &models.TWithdrawOrder{
		TenantId:     tenantId,
		UserId:       userId,
		OrderNo:      orderNo,
		BizOrderNo:   sql.NullString{String: orderNo, Valid: true},
		PlatformId:   0,
		ProductId:    0,
		AccountId:    0,
		ChannelId:    0,
		Currency:     in.Currency,
		Amount:       in.Amount,
		FeeAmount:    0,
		ActualAmount: 0,
		ClientType:   0,
		ClientIp:     sql.NullString{String: clientIP, Valid: clientIP != ""},
		Status:       int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING),
		ThirdTradeNo: sql.NullString{String: "", Valid: true},
		ThirdOrderNo: sql.NullString{String: "", Valid: true},
		RequestData:  sql.NullString{String: "{}", Valid: true},
		ResponseData: sql.NullString{String: "{}", Valid: true},
		NotifyData:   sql.NullString{String: "{}", Valid: true},
		ProcessTime:  0,
		NotifyTime:   0,
		CloseTime:    0,
		Remark:       sql.NullString{String: "用户提现", Valid: true},
		CreateTimes:  now,
		UpdateTimes:  now,
	}

	var id int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		withdrawOrderModel := models.NewTWithdrawOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.WithdrawOrderModel)

		res, err := withdrawOrderModel.Insert(ctx, withdrawOrder)
		if err != nil {
			return err
		}

		id, _ = res.LastInsertId()
		if id == 0 {
			id = withdrawOrder.Id
		}
		withdrawOrder.Id = id

		return freezeWithdrawOrderAsset(ctx, l.svcCtx, withdrawOrder, "withdraw apply freeze")
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("Create withdraw order success: %s, user_id: %d, amount: %d", orderNo, userId, in.Amount)
	event := notify.NewEvent(notify.EventTypeWithdraw, notify.EventLevelWarning, "用户提现", fmt.Sprintf("用户 %d 发起提现订单 %s", userId, orderNo))
	event.Source = "payment"
	event.TenantID = tenantId
	event.UserID = userId
	event.BizNo = orderNo
	event.Data = map[string]any{
		"amount":   in.Amount,
		"currency": in.Currency,
	}
	if err := notify.Publish(l.ctx, l.svcCtx.Redis, event); err != nil {
		l.Errorf("publish admin withdraw notification failed, orderNo=%s err=%v", orderNo, err)
	}

	return &payment.CreateWithdrawOrderResp{
		Base: helper.OkResp(),
		Id:   id,
	}, nil
}
