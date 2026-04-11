package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/trade"
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
func (l *SetUserSymbolLimitLogic) SetUserSymbolLimit(in *trade.SetUserSymbolLimitReq) (*trade.AdminCommonResp, error) {
	now := utils.NowMillis()
	item, err := l.svcCtx.RiskUserSymbolLimitModel.FindOneByTenantIdUserIdSymbolIdMarketType(l.ctx, in.TenantId, in.UserId, in.SymbolId, int64(in.MarketType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil {
		item = &models.TRiskUserSymbolLimit{TenantId: in.TenantId, UserId: in.UserId, SymbolId: in.SymbolId, MarketType: int64(in.MarketType), CreateTimes: now}
	}
	item.MaxPositionQty = mustParseFloat(in.MaxPositionQty)
	item.MaxPositionNotional = mustParseFloat(in.MaxPositionNotional)
	item.MaxOpenOrders = int64(in.MaxOpenOrders)
	item.MaxOrderQty = mustParseFloat(in.MaxOrderQty)
	item.MaxOrderNotional = mustParseFloat(in.MaxOrderNotional)
	item.MinOrderQty = mustParseFloat(in.MinOrderQty)
	item.MinOrderNotional = mustParseFloat(in.MinOrderNotional)
	item.MaxLongPositionQty = mustParseFloat(in.MaxLongPositionQty)
	item.MaxShortPositionQty = mustParseFloat(in.MaxShortPositionQty)
	item.PriceDeviationRate = mustParseFloat(in.PriceDeviationRate)
	item.OperatorId = in.OperatorId
	item.Source = int64(in.Source)
	item.Status = int64(in.Status)
	item.EffectiveStartTime = in.EffectiveStartTime
	item.EffectiveEndTime = in.EffectiveEndTime
	item.Remark = in.Remark
	item.UpdateTimes = now
	if item.Id == 0 {
		_, err = l.svcCtx.RiskUserSymbolLimitModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.RiskUserSymbolLimitModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
