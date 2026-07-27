package adminlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"wklive/services/payment/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/provider"
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
func (l *AuditWithdrawOrderLogic) AuditWithdrawOrder(in *payment.AuditWithdrawOrderReq) (*payment.CommonResp, error) {
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
		return &payment.CommonResp{
			Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx)),
		}, nil
	}
	if _, base, err := applyAdminTenantUpdateScope(l.ctx, order.TenantId, i18n.OrderNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return base, nil
	}

	// 只有待审核状态的订单才能审核
	if order.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) {
		return &payment.CommonResp{
			Base: helper.ErrResp(i18n.OnlyPendingReviewOrdersCanAudit, i18n.Translate(i18n.OnlyPendingReviewOrdersCanAudit, l.ctx)),
		}, nil
	}

	platform, err := l.svcCtx.PayPlatformModel.FindOne(l.ctx, order.PlatformId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	isThirdPartyPayout := in.Approve == 1 &&
		platform != nil &&
		platform.PlatformType == int64(payment.PlatformType_PLATFORM_TYPE_THIRD)
	if isThirdPartyPayout {
		if order.AccountId <= 0 {
			return nil, fmt.Errorf("third-party payout account is required")
		}
		account, err := l.svcCtx.TenantPayAccountModel.FindOne(l.ctx, order.AccountId)
		if err != nil {
			return nil, err
		}
		if _, err := l.svcCtx.PaymentAdapters.Get(platform.PlatformCode); err != nil {
			return nil, err
		}
		if account.PlatformId != platform.Id || account.TenantId != order.TenantId {
			return nil, fmt.Errorf("withdraw payout account does not match order platform or tenant")
		}
	}

	now := utils.NowMillis()
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		withdrawOrderModel := models.NewTWithdrawOrderModel(conn, l.svcCtx.Config.CacheRedis)

		current, err := withdrawOrderModel.FindOne(ctx, order.Id)
		if err != nil {
			return err
		}
		if current.Status != int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PENDING) {
			return i18n.StatusError(ctx, i18n.OnlyPendingReviewOrdersCanAudit)
		}

		if in.Approve == 1 {
			// 审核通过只进入处理中；三方代付确认成功后才能扣除冻结资金并标记成功。
			current.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PAYING)
			current.ProcessTime = now
		} else {
			if err := helpers.UnfreezeWithdrawOrderAsset(ctx, l.svcCtx, current, "withdraw audit rejected"); err != nil {
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

	if isThirdPartyPayout {
		if err := l.createThirdPartyPayout(platform, order); err != nil {
			l.Errorf("%s create third-party payout failed, orderNo=%s err=%v", errLogic, order.OrderNo, err)
			return nil, err
		}
	}

	l.Logger.Infof("Audit withdraw order success: %s, approve: %v", in.OrderNo, in.Approve)

	return &payment.CommonResp{
		Base: helper.OkResp(),
	}, nil
}

func (l *AuditWithdrawOrderLogic) createThirdPartyPayout(
	platform *models.TPayPlatform,
	order *models.TWithdrawOrder,
) error {
	requestNo := order.OrderNo + "_PAYOUT"
	existing, err := l.svcCtx.PayRequestLogModel.FindOneByRequestNo(l.ctx, requestNo)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if existing != nil {
		switch payment.PayRequestStatus(existing.Status) {
		case payment.PayRequestStatus_PAY_REQUEST_STATUS_SUCCESS,
			payment.PayRequestStatus_PAY_REQUEST_STATUS_PROCESSING,
			payment.PayRequestStatus_PAY_REQUEST_STATUS_UNCERTAIN:
			return nil
		}
	}

	account, err := l.svcCtx.TenantPayAccountModel.FindOne(l.ctx, order.AccountId)
	if err != nil {
		return err
	}
	adapter, err := l.svcCtx.PaymentAdapters.Get(platform.PlatformCode)
	if err != nil {
		return err
	}

	startedAt := utils.NowMillis()
	requestLog := &models.TPayRequestLog{
		TenantId: order.TenantId, OrderType: 2, OrderId: order.Id,
		OrderNo: order.OrderNo, PlatformId: order.PlatformId, AccountId: order.AccountId,
		RequestType: 5, RequestNo: requestNo,
		Status:      int64(payment.PayRequestStatus_PAY_REQUEST_STATUS_PROCESSING),
		CreateTimes: startedAt, UpdateTimes: startedAt,
	}
	if existing == nil {
		if _, err := l.svcCtx.PayRequestLogModel.Insert(l.ctx, requestLog); err != nil {
			if current, findErr := l.svcCtx.PayRequestLogModel.FindOneByRequestNo(l.ctx, requestNo); findErr == nil && current != nil {
				return nil
			}
			return err
		}
	} else {
		requestLog = existing
		requestLog.Status = int64(payment.PayRequestStatus_PAY_REQUEST_STATUS_PROCESSING)
		requestLog.UpdateTimes = startedAt
		if err := l.svcCtx.PayRequestLogModel.Update(l.ctx, requestLog); err != nil {
			return err
		}
	}

	result, payoutErr := adapter.CreatePayout(l.ctx, account, order)
	finishedAt := utils.NowMillis()
	requestLog.DurationMs = finishedAt - startedAt
	requestLog.UpdateTimes = finishedAt
	if payoutErr != nil {
		// A timeout or network failure may occur after the provider accepted the
		// payout. Keep funds frozen and resolve it through QueryPayout.
		requestLog.Status = int64(payment.PayRequestStatus_PAY_REQUEST_STATUS_UNCERTAIN)
		requestLog.ThirdMessage = payoutErr.Error()
		_ = l.svcCtx.PayRequestLogModel.Update(l.ctx, requestLog)
		return payoutErr
	}
	if result == nil {
		requestLog.Status = int64(payment.PayRequestStatus_PAY_REQUEST_STATUS_UNCERTAIN)
		requestLog.ThirdMessage = "empty payout response"
		_ = l.svcCtx.PayRequestLogModel.Update(l.ctx, requestLog)
		return fmt.Errorf("empty payout response")
	}

	responseJSON, _ := json.Marshal(result)
	requestLog.ResponseData = sql.NullString{String: string(responseJSON), Valid: true}
	requestLog.Status = int64(payment.PayRequestStatus_PAY_REQUEST_STATUS_SUCCESS)
	if err := l.svcCtx.PayRequestLogModel.Update(l.ctx, requestLog); err != nil {
		return err
	}

	return l.applyThirdPartyPayoutResult(order.Id, result, finishedAt)
}

func (l *AuditWithdrawOrderLogic) applyThirdPartyPayoutResult(
	orderID int64,
	payoutResult *provider.CreatePayoutResult,
	finishedAt int64,
) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		withdrawOrderModel := models.NewTWithdrawOrderModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := withdrawOrderModel.FindOne(ctx, orderID)
		if err != nil {
			return err
		}
		current.ThirdOrderNo = sql.NullString{String: payoutResult.ThirdOrderNo, Valid: payoutResult.ThirdOrderNo != ""}
		current.ThirdTradeNo = sql.NullString{String: payoutResult.ThirdTradeNo, Valid: payoutResult.ThirdTradeNo != ""}
		responseJSON, _ := json.Marshal(payoutResult)
		current.ResponseData = sql.NullString{String: string(responseJSON), Valid: true}
		current.UpdateTimes = finishedAt

		switch payment.PayOrderStatus(payoutResult.Status) {
		case payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS:
			if err := helpers.DeductWithdrawOrderFrozenAsset(ctx, l.svcCtx, current, "third-party payout success"); err != nil {
				return err
			}
			current.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_SUCCESS)
			current.CloseTime = finishedAt
		case payment.PayOrderStatus_PAY_ORDER_STATUS_FAILED,
			payment.PayOrderStatus_PAY_ORDER_STATUS_CLOSED:
			if err := helpers.UnfreezeWithdrawOrderAsset(ctx, l.svcCtx, current, "third-party payout failed"); err != nil {
				return err
			}
			current.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_FAILED)
			current.CloseTime = finishedAt
		default:
			current.Status = int64(payment.PayOrderStatus_PAY_ORDER_STATUS_PAYING)
		}
		return withdrawOrderModel.Update(ctx, current)
	})
}
