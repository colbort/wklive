package adminlogic

import (
	"context"
	"database/sql"
	"errors"
	"time"
	helpers "wklive/services/trade/internal/logic/helpers"

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

	now := utils.NowMillis()
	item, err := l.svcCtx.RiskUserSymbolLimitModel.FindOneByTenantIdUserIdSymbolId(l.ctx, in.TenantId, in.UserId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil {
		item = &models.TRiskUserSymbolLimit{
			TenantId:    in.TenantId,
			UserId:      in.UserId,
			SymbolId:    in.SymbolId,
			Enabled:     int64(common.Enable_ENABLE_ENABLED),
			CreateTimes: now,
		}
	}
	item.MaxPositionQty = helpers.MustParseFloat(in.MaxPositionQty)
	item.MaxPositionNotional = helpers.MustParseFloat(in.MaxPositionNotional)
	item.MaxOpenOrders = int64(in.MaxOpenOrders)
	item.MaxOrderQty = helpers.MustParseFloat(in.MaxOrderQty)
	item.MaxOrderNotional = helpers.MustParseFloat(in.MaxOrderNotional)
	item.MinOrderQty = helpers.MustParseFloat(in.MinOrderQty)
	item.MinOrderNotional = helpers.MustParseFloat(in.MinOrderNotional)
	item.MaxLongPositionQty = helpers.MustParseFloat(in.MaxLongPositionQty)
	item.MaxShortPositionQty = helpers.MustParseFloat(in.MaxShortPositionQty)
	item.PriceDeviationRate = helpers.MustParseFloat(in.PriceDeviationRate)
	item.OperatorId = in.OperatorId
	item.Source = int64(in.Source)
	item.Enabled = helpers.EnableToModel(in.Enabled, item.Enabled)
	item.EffectiveStartTime = in.EffectiveStartTime
	item.EffectiveEndTime = in.EffectiveEndTime
	item.Remark = in.Remark
	item.UpdateTimes = now
	if item.Id == 0 {
		var result sql.Result
		result, err = l.svcCtx.RiskUserSymbolLimitModel.Insert(l.ctx, item)
		if err == nil {
			item.Id, err = result.LastInsertId()
		}
	} else {
		err = l.svcCtx.RiskUserSymbolLimitModel.Update(l.ctx, item)
	}
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
