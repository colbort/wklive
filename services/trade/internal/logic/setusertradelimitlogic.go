package logic

import (
	"context"
	"errors"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/trade"
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
func (l *SetUserTradeLimitLogic) SetUserTradeLimit(in *trade.SetUserTradeLimitReq) (*trade.AdminCommonResp, error) {
	now := utils.NowMillis()
	item, err := l.svcCtx.RiskUserTradeLimitModel.FindOneByTenantIdUserIdMarketType(l.ctx, in.TenantId, in.UserId, int64(in.MarketType))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil {
		item = &models.TRiskUserTradeLimit{TenantId: in.TenantId, UserId: in.UserId, MarketType: int64(in.MarketType), CreateTimes: now}
	}
	item.CanOpen = conv.BoolInt64(in.CanOpen)
	item.CanClose = conv.BoolInt64(in.CanClose)
	item.CanCancel = conv.BoolInt64(in.CanCancel)
	item.CanTriggerOrder = conv.BoolInt64(in.CanTriggerOrder)
	item.CanApiTrade = conv.BoolInt64(in.CanApiTrade)
	item.TradeEnabled = conv.BoolInt64(in.TradeEnabled)
	item.OnlyReduceOnly = conv.BoolInt64(in.OnlyReduceOnly)
	item.MaxOpenOrderCount = int64(in.MaxOpenOrderCount)
	item.MaxOrderCountPerDay = int64(in.MaxOrderCountPerDay)
	item.MaxCancelCountPerDay = int64(in.MaxCancelCountPerDay)
	item.MaxOpenNotional = mustParseFloat(in.MaxOpenNotional)
	item.MaxPositionNotional = mustParseFloat(in.MaxPositionNotional)
	item.RiskLevel = int64(in.RiskLevel)
	item.OperatorId = in.OperatorId
	item.Source = int64(in.Source)
	item.Status = int64(in.Status)
	item.EffectiveStartTime = in.EffectiveStartTime
	item.EffectiveEndTime = in.EffectiveEndTime
	item.Remark = in.Remark
	item.UpdateTimes = now
	if item.Id == 0 {
		_, err = l.svcCtx.RiskUserTradeLimitModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.RiskUserTradeLimitModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
