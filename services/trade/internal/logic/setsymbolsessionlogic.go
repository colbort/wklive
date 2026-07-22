package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/authz"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetSymbolSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetSymbolSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetSymbolSessionLogic {
	return &SetSymbolSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存交易时段配置
func (l *SetSymbolSessionLogic) SetSymbolSession(in *trade.SetSymbolSessionReq) (*trade.AdminCommonResp, error) {
	if in.DayOfWeek < 1 || in.DayOfWeek > 7 || in.StartSecond < 0 || in.EndSecond <= in.StartSecond || in.EndSecond > 86400 || in.Timezone == "" {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	symbol, err := l.svcCtx.TradeSymbolModel.FindOne(l.ctx, in.SymbolId)
	if errors.Is(err, models.ErrNotFound) {
		return &trade.AdminCommonResp{Base: helper.ErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, symbol.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.AdminCommonResp{Base: base}, nil
	}
	item, err := l.svcCtx.TradeSymbolSessionModel.FindOneByTenantIdSymbolIdDayOfWeekStartSecondEndSecond(l.ctx, symbol.TenantId, in.SymbolId, in.DayOfWeek, in.StartSecond, in.EndSecond)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil {
		item = &models.TTradeSymbolSession{TenantId: symbol.TenantId, SymbolId: in.SymbolId, DayOfWeek: in.DayOfWeek, StartSecond: in.StartSecond, EndSecond: in.EndSecond}
		item.CreateTimes = utils.NowMillis()
	}
	item.Timezone = in.Timezone
	item.Enabled = enableToModel(in.Enabled, item.Enabled)
	if item.Enabled == 0 {
		item.Enabled = 1
	}
	item.UpdateTimes = utils.NowMillis()
	if item.Id == 0 {
		_, err = l.svcCtx.TradeSymbolSessionModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.TradeSymbolSessionModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.AdminCommonResp{Base: helper.OkResp()}, nil
}
