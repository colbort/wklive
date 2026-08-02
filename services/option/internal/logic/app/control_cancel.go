package applogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ControlCancelAudit struct {
	EventType  string
	OperatorID int64
	Detail     string
}

// CancelOrderByControl cancels one active order using the same freeze-release
// semantics as a user cancellation. It is idempotent for terminal orders and
// is shared by kill switch, circuit breaker and administrative controls.
func CancelOrderByControl(
	ctx context.Context, svcCtx *svc.ServiceContext, orderID int64, reason string,
) (*models.TOptionOrder, error) {
	return retryControlCancellation(ctx, svcCtx, orderID, reason, nil)
}

// CancelOrderByControlWithAudit atomically appends an immutable control event
// when an active order is canceled. Terminal-order replays are no-ops and do
// not append duplicate audit evidence.
func CancelOrderByControlWithAudit(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	orderID int64,
	reason string,
	audit ControlCancelAudit,
) (*models.TOptionOrder, error) {
	audit.EventType = strings.TrimSpace(audit.EventType)
	if audit.EventType == "" || audit.OperatorID <= 0 {
		return nil, errors.New("control cancellation audit requires event type and operator")
	}
	return retryControlCancellation(ctx, svcCtx, orderID, reason, &audit)
}

const controlCancelMaxAttempts = 5

func retryControlCancellation(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	orderID int64,
	reason string,
	audit *ControlCancelAudit,
) (*models.TOptionOrder, error) {
	var lastErr error
	for attempt := 0; attempt < controlCancelMaxAttempts; attempt++ {
		order, err := cancelOrderByControl(ctx, svcCtx, orderID, reason, audit)
		if err == nil {
			return order, nil
		}
		lastErr = err
		if !isRetryableControlCancelError(err) || attempt == controlCancelMaxAttempts-1 {
			return nil, err
		}
		delay := time.Duration(10*(1<<attempt)) * time.Millisecond
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, errors.Join(ctx.Err(), lastErr)
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func isRetryableControlCancelError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && (mysqlErr.Number == 1213 || mysqlErr.Number == 1205)
}

func cancelOrderByControl(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	orderID int64,
	reason string,
	audit *ControlCancelAudit,
) (*models.TOptionOrder, error) {
	initial, err := svcCtx.OptionOrderModel.FindOne(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if initial.ComboOrderId > 0 {
		if err := CancelComboOrderByControl(
			ctx, svcCtx, initial.ComboOrderId, reason,
		); err != nil {
			return nil, err
		}
		return svcCtx.OptionOrderModel.FindOne(ctx, orderID)
	}
	var canceled *models.TOptionOrder
	err = svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, svcCtx.Config.CacheRedis)
		order, err := orderModel.FindOneForUpdate(txCtx, orderID)
		if err != nil {
			return err
		}
		switch option.OrderStatus(order.Status) {
		case option.OrderStatus_ORDER_STATUS_FUNDING,
			option.OrderStatus_ORDER_STATUS_PENDING,
			option.OrderStatus_ORDER_STATUS_PART_FILLED:
		default:
			return nil
		}
		now := time.Now().Unix()
		cancelBeforeFreeze := false
		if order.Status == int64(option.OrderStatus_ORDER_STATUS_FUNDING) {
			freeze, err := instructionModel.FindOneByTenantIdInstructionNo(
				txCtx, order.TenantId, order.OrderNo+"-FREEZE",
			)
			if err != nil {
				return err
			}
			freeze, err = instructionModel.FindOneForUpdate(txCtx, freeze.Id)
			if err != nil {
				return err
			}
			if freeze.Status == int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING) {
				freeze.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_CANCELED)
				freeze.UpdateTimes = now
				if err := instructionModel.Update(txCtx, freeze); err != nil {
					return err
				}
				cancelBeforeFreeze = true
				order.MarginAmount = decimal.Zero
			}
		}
		if err := releaseClosePositionFrozenQty(
			txCtx, positionModel, order, order.UnfilledQty, now,
		); err != nil {
			return err
		}
		order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
		if order.MarginAmount.IsPositive() && !cancelBeforeFreeze {
			order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELING)
			if _, err := instructionModel.Insert(txCtx, &models.TOptionAssetInstruction{
				TenantId: order.TenantId, InstructionNo: order.OrderNo + "-CONTROL-RELEASE",
				BizNo: order.OrderNo, OrderId: order.Id, UserId: order.UserId, AccountId: order.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: order.OrderNo, Coin: OptionOrderMarginCoin(order), Amount: order.MarginAmount,
				StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		order.CancelReason = reason
		order.CancelTime = now
		order.UpdateTimes = now
		if err := orderModel.Update(txCtx, order); err != nil {
			return err
		}
		if audit != nil {
			eventModel := models.NewTOptionTradingControlEventModel(
				conn, svcCtx.Config.CacheRedis,
			)
			if _, err := eventModel.Insert(txCtx, &models.TOptionTradingControlEvent{
				TenantId: order.TenantId, UserId: order.UserId,
				ContractId: order.ContractId, OrderId: order.Id,
				EventType: audit.EventType, Reason: reason, Detail: audit.Detail,
				OperatorId: audit.OperatorID, CreateTimes: now,
			}); err != nil {
				return err
			}
		}
		canceled = order
		return nil
	})
	if err == nil && canceled != nil {
		PublishOptionOrderChanged(ctx, svcCtx, canceled)
	}
	return canceled, err
}

func CancelContractOrdersByControl(
	ctx context.Context, svcCtx *svc.ServiceContext, tenantID, contractID int64, reason string,
) error {
	_, _, _, err := CancelContractOrdersByControlReport(
		ctx, svcCtx, tenantID, contractID, reason, false,
	)
	return err
}

// CancelContractOrdersByControlReport cancels every active order visible in a
// contract and returns exact per-order progress. continueOnError is used by
// emergency controls: one malformed or temporarily unavailable order must not
// prevent later orders in the batch from being made safe.
func CancelContractOrdersByControlReport(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	tenantID, contractID int64,
	reason string,
	continueOnError bool,
) (total, success, failed int64, lastErr error) {
	cursor := int64(0)
	for {
		orders, _, err := svcCtx.OptionOrderModel.FindPage(ctx, models.OptionOrderPageFilter{
			TenantId: tenantID, ContractId: contractID,
			Statuses: []int64{
				int64(option.OrderStatus_ORDER_STATUS_FUNDING),
				int64(option.OrderStatus_ORDER_STATUS_PENDING),
				int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			},
		}, cursor, 100)
		if err != nil {
			return total, success, failed, err
		}
		for _, order := range orders {
			cursor = order.Id
			total++
			if _, err := CancelOrderByControl(ctx, svcCtx, order.Id, reason); err != nil {
				failed++
				lastErr = err
				if !continueOnError {
					return total, success, failed, err
				}
				continue
			}
			success++
		}
		if len(orders) < 100 {
			return total, success, failed, lastErr
		}
	}
}
