package logic

import (
	"context"
	"database/sql"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
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
	now := utils.NowMillis()
	orderNo, err := l.svcCtx.GenerateOrderNo(l.ctx, "WD")
	if err != nil {
		return nil, err
	}

	// Create withdraw order
	withdrawOrder := &models.TWithdrawOrder{
		TenantId:     in.TenantId,
		UserId:       in.UserId,
		OrderNo:      orderNo,
		BizOrderNo:   sql.NullString{String: "", Valid: true},
		PlatformId:   0,
		ProductId:    0,
		AccountId:    0,
		ChannelId:    0,
		Currency:     in.Currency,
		Amount:       in.Amount,
		FeeAmount:    0,
		ActualAmount: 0,
		ClientType:   0,
		ClientIp:     sql.NullString{String: "", Valid: true},
		Status:       int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING),
		ThirdTradeNo: sql.NullString{String: "", Valid: true},
		ThirdOrderNo: sql.NullString{String: "", Valid: true},
		RequestData:  sql.NullString{String: "", Valid: true},
		ResponseData: sql.NullString{String: "", Valid: true},
		NotifyData:   sql.NullString{String: "", Valid: true},
		ProcessTime:  sql.NullInt64{Int64: 0, Valid: true},
		NotifyTime:   sql.NullInt64{Int64: 0, Valid: true},
		CloseTime:    sql.NullInt64{Int64: 0, Valid: true},
		Remark:       sql.NullString{String: "用户提现", Valid: true},
		CreateTimes:  now,
		UpdateTimes:  now,
	}

	res, err := l.svcCtx.WithdrawOrderModel.Insert(l.ctx, withdrawOrder)
	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()
	if id == 0 {
		id = withdrawOrder.Id
	}

	l.Logger.Infof("Create withdraw order success: %s, user_id: %d, amount: %d", orderNo, in.UserId, in.Amount)

	return &payment.CreateWithdrawOrderResp{
		Base: helper.OkResp(),
		Id:   id,
	}, nil
}
