package adminlogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/authz"
	"wklive/services/trade/internal/delayqueue"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetUserSymbolLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserSymbolLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserSymbolLimitLogic {
	return &SetUserSymbolLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户交易对限制
func (l *SetUserSymbolLimitLogic) SetUserSymbolLimit(in *trade.SetUserSymbolLimitReq) (*trade.CommonResp, error) {
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.CommonResp{Base: base}, nil
	}

	if in.TenantId <= 0 || in.UserId <= 0 || in.SymbolId <= 0 {
		return nil, errors.New("invalid user symbol control scope")
	}
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if err != nil || symbol.TenantId != in.TenantId {
		return nil, errors.New("symbol does not belong to tenant")
	}
	mode := in.ControlMode
	if mode == trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_UNKNOWN {
		mode = trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_NORMAL
	}
	if mode < trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_NORMAL || mode > trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_DISABLED {
		return nil, errors.New("invalid user symbol control mode")
	}
	if in.EffectiveEndTime > 0 && in.EffectiveStartTime > 0 && in.EffectiveEndTime <= in.EffectiveStartTime {
		return nil, errors.New("effective end time must be after start time")
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	var item *models.TRiskUserSymbolLimit
	err = l.svcCtx.TransactionModel.TransactOnce(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		current, findErr := tx.RiskUserSymbolLimit.FindOneByTenantIdUserIdSymbolId(ctx, in.TenantId, in.UserId, in.SymbolId)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		changeType := int64(2)
		before := sql.NullString{}
		if current == nil {
			changeType = 1
			current = &models.TRiskUserSymbolLimit{
				TenantId: in.TenantId, UserId: in.UserId, SymbolId: in.SymbolId,
				ControlMode: int64(mode), Enabled: int64(common.Enable_ENABLE_ENABLED), Version: 1, CreateTimes: now,
			}
		} else {
			if in.ExpectedVersion > 0 && current.Version != in.ExpectedVersion {
				return fmt.Errorf("user symbol control was changed concurrently: expected version %d, current %d", in.ExpectedVersion, current.Version)
			}
			raw, _ := json.Marshal(current)
			before = sql.NullString{String: string(raw), Valid: true}
			current.Version++
		}
		current.ControlMode = int64(mode)
		current.MaxPositionQty = helpers.MustParseFloat(in.MaxPositionQty)
		current.MaxPositionNotional = helpers.MustParseFloat(in.MaxPositionNotional)
		current.MaxOpenOrders = int64(in.MaxOpenOrders)
		current.MaxOrderQty = helpers.MustParseFloat(in.MaxOrderQty)
		current.MaxOrderNotional = helpers.MustParseFloat(in.MaxOrderNotional)
		current.MinOrderQty = helpers.MustParseFloat(in.MinOrderQty)
		current.MinOrderNotional = helpers.MustParseFloat(in.MinOrderNotional)
		current.MaxLongPositionQty = helpers.MustParseFloat(in.MaxLongPositionQty)
		current.MaxShortPositionQty = helpers.MustParseFloat(in.MaxShortPositionQty)
		current.PriceDeviationRate = helpers.MustParseFloat(in.PriceDeviationRate)
		current.OperatorId = operatorID
		current.Source = int64(trade.SourceType_SOURCE_TYPE_ADMIN)
		current.Enabled = helpers.EnableToModel(in.Enabled, current.Enabled)
		current.EffectiveStartTime = in.EffectiveStartTime
		current.EffectiveEndTime = in.EffectiveEndTime
		current.Remark = in.Remark
		current.UpdateTimes = now
		if current.Id == 0 {
			var result sql.Result
			result, findErr = tx.RiskUserSymbolLimit.Insert(ctx, current)
			if findErr == nil {
				current.Id, findErr = result.LastInsertId()
			}
		} else {
			findErr = tx.RiskUserSymbolLimit.Update(ctx, current)
		}
		if findErr != nil {
			return findErr
		}
		afterRaw, _ := json.Marshal(current)
		_, findErr = tx.TradeUserControlAudit.Insert(ctx, &models.TTradeUserControlAudit{
			TenantId: current.TenantId, ControlId: current.Id, UserId: current.UserId,
			ScopeType:  2,
			ChangeType: changeType, BeforeJson: before,
			AfterJson:  sql.NullString{String: string(afterRaw), Valid: true},
			OperatorId: operatorID, Source: int64(trade.SourceType_SOURCE_TYPE_ADMIN),
			Reason: current.Remark, CreateTimes: now,
		})
		item = current
		return findErr
	})
	if err != nil {
		return nil, err
	}
	if item.Enabled == int64(common.Enable_ENABLE_ENABLED) && item.EffectiveEndTime > now && l.svcCtx.DelayQueue != nil {
		if err := l.svcCtx.DelayQueue.At(delayqueue.Message{
			Action: delayqueue.ActionExpireSymbolRisk, TenantID: item.TenantId,
			EntityID: item.Id, DueAt: item.EffectiveEndTime,
		}, time.UnixMilli(item.EffectiveEndTime)); err != nil {
			l.Errorf("enqueue symbol risk expiration failed, id=%d err=%v", item.Id, err)
		}
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
