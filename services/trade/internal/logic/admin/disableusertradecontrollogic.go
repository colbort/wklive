package adminlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/authz"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type DisableUserTradeControlLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDisableUserTradeControlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableUserTradeControlLogic {
	return &DisableUserTradeControlLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 停用用户交易控制
func (l *DisableUserTradeControlLogic) DisableUserTradeControl(in *trade.DisableUserTradeControlReq) (*trade.CommonResp, error) {
	reason := strings.TrimSpace(in.Reason)
	if in.ControlId <= 0 || (in.ScopeType != 1 && in.ScopeType != 2) || in.ExpectedVersion <= 0 || reason == "" {
		return nil, errors.New("control id, scope type, expected version and reason are required")
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	err = l.svcCtx.TransactionModel.TransactOnce(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		if in.ScopeType == 1 {
			item, findErr := tx.RiskUserTradeLimit.FindOneForUpdate(ctx, in.ControlId)
			if findErr != nil {
				return findErr
			}
			if in.TenantId > 0 && item.TenantId != in.TenantId {
				return errors.New("control does not belong to requested tenant")
			}
			if base, scopeErr := authz.AdminTenantWriteScopeResp(ctx, item.TenantId, i18n.BusinessDataNotFound); scopeErr != nil {
				return scopeErr
			} else if base != nil {
				return fmt.Errorf("permission denied: %s", base.Msg)
			}
			if item.Version != in.ExpectedVersion {
				return fmt.Errorf("user trade control was changed concurrently: expected version %d, current %d", in.ExpectedVersion, item.Version)
			}
			if item.Enabled == int64(common.Enable_ENABLE_DISABLED) {
				return nil
			}
			beforeRaw, _ := json.Marshal(item)
			item.Enabled = int64(common.Enable_ENABLE_DISABLED)
			item.Version++
			item.OperatorId = operatorID
			item.Source = int64(trade.SourceType_SOURCE_TYPE_ADMIN)
			item.Remark = reason
			item.UpdateTimes = now
			if findErr = tx.RiskUserTradeLimit.Update(ctx, item); findErr != nil {
				return findErr
			}
			afterRaw, _ := json.Marshal(item)
			_, findErr = tx.TradeUserControlAudit.Insert(ctx, &models.TTradeUserControlAudit{
				TenantId: item.TenantId, ControlId: item.Id, ScopeType: 1, UserId: item.UserId, ChangeType: 3,
				BeforeJson: sql.NullString{String: string(beforeRaw), Valid: true}, AfterJson: sql.NullString{String: string(afterRaw), Valid: true},
				OperatorId: operatorID, Source: int64(trade.SourceType_SOURCE_TYPE_ADMIN), Reason: reason, CreateTimes: now,
			})
			return findErr
		}
		item, findErr := tx.RiskUserSymbolLimit.FindOneForUpdate(ctx, in.ControlId)
		if findErr != nil {
			return findErr
		}
		if in.TenantId > 0 && item.TenantId != in.TenantId {
			return errors.New("control does not belong to requested tenant")
		}
		if base, scopeErr := authz.AdminTenantWriteScopeResp(ctx, item.TenantId, i18n.BusinessDataNotFound); scopeErr != nil {
			return scopeErr
		} else if base != nil {
			return fmt.Errorf("permission denied: %s", base.Msg)
		}
		if item.Version != in.ExpectedVersion {
			return fmt.Errorf("user symbol control was changed concurrently: expected version %d, current %d", in.ExpectedVersion, item.Version)
		}
		if item.Enabled == int64(common.Enable_ENABLE_DISABLED) {
			return nil
		}
		beforeRaw, _ := json.Marshal(item)
		item.Enabled = int64(common.Enable_ENABLE_DISABLED)
		item.Version++
		item.OperatorId = operatorID
		item.Source = int64(trade.SourceType_SOURCE_TYPE_ADMIN)
		item.Remark = reason
		item.UpdateTimes = now
		if findErr = tx.RiskUserSymbolLimit.Update(ctx, item); findErr != nil {
			return findErr
		}
		afterRaw, _ := json.Marshal(item)
		_, findErr = tx.TradeUserControlAudit.Insert(ctx, &models.TTradeUserControlAudit{
			TenantId: item.TenantId, ControlId: item.Id, ScopeType: 2, UserId: item.UserId, ChangeType: 3,
			BeforeJson: sql.NullString{String: string(beforeRaw), Valid: true}, AfterJson: sql.NullString{String: string(afterRaw), Valid: true},
			OperatorId: operatorID, Source: int64(trade.SourceType_SOURCE_TYPE_ADMIN), Reason: reason, CreateTimes: now,
		})
		return findErr
	})
	if err != nil {
		return nil, err
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
