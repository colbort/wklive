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

type SetUserTradeLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetUserTradeLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserTradeLimitLogic {
	return &SetUserTradeLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户交易限制
func (l *SetUserTradeLimitLogic) SetUserTradeLimit(in *trade.SetUserTradeLimitReq) (*trade.CommonResp, error) {
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.CommonResp{Base: base}, nil
	}

	if in.TenantId <= 0 || in.UserId <= 0 || in.ProductType < common.ProductType_PRODUCT_TYPE_SPOT || in.ProductType > common.ProductType_PRODUCT_TYPE_SECONDS {
		return nil, errors.New("invalid user trade control scope")
	}
	contractType := int64(in.ContractType)
	if in.ProductType != common.ProductType_PRODUCT_TYPE_DERIVATIVE {
		contractType = 0
	} else if contractType < 0 || contractType > int64(common.ContractType_CONTRACT_TYPE_DELIVERY) {
		return nil, errors.New("invalid contract type")
	}
	mode := in.ControlMode
	if mode == trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_UNKNOWN {
		mode = trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_NORMAL
	}
	if mode < trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_NORMAL || mode > trade.UserTradeControlMode_USER_TRADE_CONTROL_MODE_DISABLED {
		return nil, errors.New("invalid user trade control mode")
	}
	if in.EffectiveEndTime > 0 && in.EffectiveStartTime > 0 && in.EffectiveEndTime <= in.EffectiveStartTime {
		return nil, errors.New("effective end time must be after start time")
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	var item *models.TRiskUserTradeLimit
	err = l.svcCtx.TransactionModel.TransactOnce(l.ctx, func(ctx context.Context, tx *models.TransactionModels) error {
		current, findErr := tx.RiskUserTradeLimit.FindOneByTenantIdUserIdProductTypeContractType(ctx, in.TenantId, in.UserId, int64(in.ProductType), contractType)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		changeType := int64(2)
		before := sql.NullString{}
		if current == nil {
			changeType = 1
			current = &models.TRiskUserTradeLimit{
				TenantId: in.TenantId, UserId: in.UserId, ProductType: int64(in.ProductType), ContractType: contractType,
				TradeEnabled: int64(common.Enable_ENABLE_ENABLED), OnlyReduceOnly: int64(common.Enable_ENABLE_DISABLED),
				Enabled: int64(common.Enable_ENABLE_ENABLED), Version: 1, CreateTimes: now,
			}
		} else {
			if in.ExpectedVersion > 0 && current.Version != in.ExpectedVersion {
				return fmt.Errorf("user trade control was changed concurrently: expected version %d, current %d", in.ExpectedVersion, current.Version)
			}
			raw, _ := json.Marshal(current)
			before = sql.NullString{String: string(raw), Valid: true}
			current.Version++
		}
		current.ControlMode = int64(mode)
		current.CanOpen = in.CanOpen
		current.CanClose = in.CanClose
		current.CanCancel = in.CanCancel
		current.CanTriggerOrder = in.CanTriggerOrder
		current.CanApiTrade = in.CanApiTrade
		current.TradeEnabled = helpers.EnableToModel(in.TradeEnabled, current.TradeEnabled)
		current.OnlyReduceOnly = helpers.EnableToModel(common.Enable(in.OnlyReduceOnly), current.OnlyReduceOnly)
		current.MaxOpenOrderCount = in.MaxOpenOrderCount
		current.MaxOrderCountPerDay = in.MaxOrderCountPerDay
		current.MaxCancelCountPerDay = in.MaxCancelCountPerDay
		current.MaxOpenNotional = helpers.MustParseFloat(in.MaxOpenNotional)
		current.MaxPositionNotional = helpers.MustParseFloat(in.MaxPositionNotional)
		current.RiskLevel = in.RiskLevel
		current.OperatorId = operatorID
		current.Source = int64(trade.SourceType_SOURCE_TYPE_ADMIN)
		current.Enabled = helpers.EnableToModel(in.Enabled, current.Enabled)
		current.EffectiveStartTime = in.EffectiveStartTime
		current.EffectiveEndTime = in.EffectiveEndTime
		current.Remark = in.Remark
		current.UpdateTimes = now
		if current.Id == 0 {
			var result sql.Result
			result, findErr = tx.RiskUserTradeLimit.Insert(ctx, current)
			if findErr == nil {
				current.Id, findErr = result.LastInsertId()
			}
		} else {
			findErr = tx.RiskUserTradeLimit.Update(ctx, current)
		}
		if findErr != nil {
			return findErr
		}
		afterRaw, _ := json.Marshal(current)
		_, findErr = tx.TradeUserControlAudit.Insert(ctx, &models.TTradeUserControlAudit{
			TenantId: current.TenantId, ControlId: current.Id, UserId: current.UserId,
			ScopeType:  1,
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
			Action: delayqueue.ActionExpireTradeRisk, TenantID: item.TenantId,
			EntityID: item.Id, DueAt: item.EffectiveEndTime,
		}, time.UnixMilli(item.EffectiveEndTime)); err != nil {
			l.Errorf("enqueue trade risk expiration failed, id=%d err=%v", item.Id, err)
		}
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}
