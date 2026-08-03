package tasklogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"
)

func expireProductControl(ctx context.Context, svcCtx *svc.ServiceContext, id, tenantID, dueAt, now int64) (bool, error) {
	expired := false
	err := svcCtx.TransactionModel.TransactOnce(ctx, func(txCtx context.Context, tx *models.TransactionModels) error {
		item, err := tx.RiskUserTradeLimit.FindOneForUpdate(txCtx, id)
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if (tenantID > 0 && item.TenantId != tenantID) || item.Enabled != int64(common.Enable_ENABLE_ENABLED) ||
			item.EffectiveEndTime <= 0 || item.EffectiveEndTime > now || (dueAt > 0 && item.EffectiveEndTime != dueAt) {
			return nil
		}
		beforeRaw, _ := json.Marshal(item)
		item.Enabled = int64(common.Enable_ENABLE_DISABLED)
		item.Version++
		item.OperatorId = 0
		item.Source = int64(trade.SourceType_SOURCE_TYPE_TASK)
		item.UpdateTimes = now
		if err = tx.RiskUserTradeLimit.Update(txCtx, item); err != nil {
			return err
		}
		afterRaw, _ := json.Marshal(item)
		_, err = tx.TradeUserControlAudit.Insert(txCtx, &models.TTradeUserControlAudit{
			TenantId: item.TenantId, ControlId: item.Id, ScopeType: 1, UserId: item.UserId, ChangeType: 4,
			BeforeJson: sql.NullString{String: string(beforeRaw), Valid: true}, AfterJson: sql.NullString{String: string(afterRaw), Valid: true},
			Source: int64(trade.SourceType_SOURCE_TYPE_TASK), Reason: "expired automatically", CreateTimes: now,
		})
		expired = err == nil
		return err
	})
	return expired, err
}

func expireSymbolControl(ctx context.Context, svcCtx *svc.ServiceContext, id, tenantID, dueAt, now int64) (bool, error) {
	expired := false
	err := svcCtx.TransactionModel.TransactOnce(ctx, func(txCtx context.Context, tx *models.TransactionModels) error {
		item, err := tx.RiskUserSymbolLimit.FindOneForUpdate(txCtx, id)
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if (tenantID > 0 && item.TenantId != tenantID) || item.Enabled != int64(common.Enable_ENABLE_ENABLED) ||
			item.EffectiveEndTime <= 0 || item.EffectiveEndTime > now || (dueAt > 0 && item.EffectiveEndTime != dueAt) {
			return nil
		}
		beforeRaw, _ := json.Marshal(item)
		item.Enabled = int64(common.Enable_ENABLE_DISABLED)
		item.Version++
		item.OperatorId = 0
		item.Source = int64(trade.SourceType_SOURCE_TYPE_TASK)
		item.UpdateTimes = now
		if err = tx.RiskUserSymbolLimit.Update(txCtx, item); err != nil {
			return err
		}
		afterRaw, _ := json.Marshal(item)
		_, err = tx.TradeUserControlAudit.Insert(txCtx, &models.TTradeUserControlAudit{
			TenantId: item.TenantId, ControlId: item.Id, ScopeType: 2, UserId: item.UserId, ChangeType: 4,
			BeforeJson: sql.NullString{String: string(beforeRaw), Valid: true}, AfterJson: sql.NullString{String: string(afterRaw), Valid: true},
			Source: int64(trade.SourceType_SOURCE_TYPE_TASK), Reason: "expired automatically", CreateTimes: now,
		})
		expired = err == nil
		return err
	})
	return expired, err
}

func expiryNow() int64 {
	return utils.NowMillis()
}
