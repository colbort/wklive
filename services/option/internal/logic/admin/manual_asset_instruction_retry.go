package adminlogic

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const assetInstructionManualRetryEventType = "ASSET_INSTRUCTION_MANUAL_RETRY"

func validAssetInstructionManualRetryReason(reason string) bool {
	return reason != "" && utf8.RuneCountInString(reason) <= 64
}

func resetAssetInstructionWithAudit(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	instructionID, tenantID, contractID, operatorID int64,
	reason string,
) (bool, error) {
	reset := false
	err := svcCtx.DB.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, svcCtx.Config.CacheRedis)
		locked, err := instructionModel.FindOneForUpdate(ctx, instructionID)
		if err != nil {
			return err
		}
		if locked.TenantId != tenantID || locked.DeliveryUnitId > 0 {
			return nil
		}
		if locked.Status != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED) &&
			locked.Status != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW) {
			return nil
		}

		if contractID == 0 && locked.PositionId > 0 {
			position, findErr := models.NewTOptionPositionModel(
				conn, svcCtx.Config.CacheRedis,
			).FindOne(ctx, locked.PositionId)
			if findErr != nil {
				return findErr
			}
			contractID = position.ContractId
		}
		if contractID == 0 && locked.OrderId > 0 {
			order, findErr := models.NewTOptionOrderModel(
				conn, svcCtx.Config.CacheRedis,
			).FindOne(ctx, locked.OrderId)
			if findErr != nil {
				return findErr
			}
			contractID = order.ContractId
		}
		if contractID == 0 && locked.MarginLotId > 0 {
			lot, findErr := models.NewTOptionMarginLotModel(
				conn, svcCtx.Config.CacheRedis,
			).FindOne(ctx, locked.MarginLotId)
			if findErr != nil {
				return findErr
			}
			contractID = lot.ContractId
		}

		fromStatus := locked.Status
		fromRetryCount := locked.RetryCount
		lastError := locked.LastErrorMsg
		now := time.Now().Unix()
		changed, resetErr := instructionModel.ResetForManualRetry(ctx, locked.Id, now)
		if resetErr != nil {
			return resetErr
		}
		if !changed {
			return nil
		}
		detail := fmt.Sprintf(
			"instructionId=%d instructionNo=%s bizNo=%s targetBizNo=%s action=%d coin=%s amount=%s fromStatus=%d fromRetryCount=%d lastError=%q",
			locked.Id, locked.InstructionNo, locked.BizNo, locked.TargetBizNo,
			locked.Action, locked.Coin, locked.Amount, fromStatus, fromRetryCount, lastError,
		)
		detail = truncateRunes(detail, 1000)
		if eventErr := models.InsertOptionTradingControlEvent(ctx, conn, &models.TOptionTradingControlEvent{
			TenantId: locked.TenantId, UserId: locked.UserId, ContractId: contractID,
			OrderId: locked.OrderId, EventType: assetInstructionManualRetryEventType,
			Reason: reason, Detail: detail, OperatorId: operatorID, CreateTimes: now,
		}); eventErr != nil {
			return eventErr
		}
		reset = true
		return nil
	})
	return reset, err
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
