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

type AuditWithdrawOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuditWithdrawOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditWithdrawOrderLogic {
	return &AuditWithdrawOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 审核提现订单
func (l *AuditWithdrawOrderLogic) AuditWithdrawOrder(in *payment.AuditWithdrawOrderReq) (*payment.AdminCommonResp, error) {
	var (
		errLogic = "AuditWithdrawOrder"
	)

	// 查询订单是否存在
	order, err := l.svcCtx.WithdrawOrderModel.FindOneByOrderNo(l.ctx, in.OrderNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	if order == nil {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(404, "订单不存在"),
		}, nil
	}

	// 只有待审核状态的订单才能审核
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(201, "只有待审核订单才能审核"),
		}, nil
	}

	now := time.Now().UnixMilli()
	if in.Approve == 1 {
		// 审核通过，改为已批准
		order.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS)
	} else {
		// 审核不通过，改为已拒绝
		order.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_CLOSED)
		order.Remark = sql.NullString{String: in.Remark, Valid: in.Remark != ""}
	}
	order.UpdateTimes = now

	err = l.svcCtx.WithdrawOrderModel.Update(l.ctx, order)
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Audit withdraw order success: %s, approve: %v", in.OrderNo, in.Approve)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
