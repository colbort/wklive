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
			Base: helper.GetErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}

	// 只有待审核状态的订单才能审核
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) {
		return &payment.AdminCommonResp{
			Base: helper.GetErrResp(i18n.OnlyPendingReviewOrdersCanAudit, i18n.Translate(i18n.OnlyPendingReviewOrdersCanAudit, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		withdrawOrderModel := models.NewTWithdrawOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.WithdrawOrderModel)

		current, err := withdrawOrderModel.FindOne(ctx, order.Id)
		if err != nil {
			return err
		}
		if current.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) {
			return i18n.StatusError(ctx, i18n.OnlyPendingReviewOrdersCanAudit)
		}

		if in.Approve == 1 {
			if err := deductWithdrawOrderFrozenAsset(ctx, l.svcCtx, current, "withdraw audit approved"); err != nil {
				return err
			}
			// 审核通过，改为已批准
			current.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS)
		} else {
			if err := unfreezeWithdrawOrderAsset(ctx, l.svcCtx, current, "withdraw audit rejected"); err != nil {
				return err
			}
			// 审核不通过，改为已拒绝
			current.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_CLOSED)
			current.Remark = sql.NullString{String: in.Remark, Valid: in.Remark != ""}
		}
		current.UpdateTimes = now
		return withdrawOrderModel.Update(ctx, current)
	})
	if err != nil {
		l.Logger.Errorf("%s error: %s", errLogic, err.Error())
		return nil, err
	}

	l.Logger.Infof("Audit withdraw order success: %s, approve: %v", in.OrderNo, in.Approve)

	return &payment.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
