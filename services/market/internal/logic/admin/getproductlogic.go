package adminlogic

import (
	"context"
	"errors"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取产品详情
func (l *GetProductLogic) GetProduct(in *market.GetProductReq) (*market.GetProductResp, error) {
	result, err := l.svcCtx.MarketProductModel.FindOne(l.ctx, in.Id)
	if err != nil && errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if result == nil {
		return &market.GetProductResp{
			Base: helper.ErrResp(i18n.NotFound, i18n.Translate(i18n.NotFound, l.ctx)),
		}, nil
	}
	return &market.GetProductResp{
		Base: helper.OkResp(),
		Data: productWithTradingCalendar(l.ctx, l.svcCtx, result),
	}, nil
}

func productWithTradingCalendar(ctx context.Context, svcCtx *svc.ServiceContext, product *models.TItickProduct) *market.MarketProduct {
	out := helpers.ToProductProto(product)
	if svcCtx.MarketCalendarResolver == nil {
		return out
	}
	definition := svcCtx.MarketCalendarResolver.ResolveProduct(
		ctx, product.Id, product.CategoryCode, product.Market, product.Symbol, product.Exchange,
	)
	if definition == nil || definition.ID <= 0 {
		return out
	}
	calendar := &market.MarketTradingCalendar{
		Id: definition.ID, CategoryCode: definition.CategoryCode, Market: definition.Market,
		Exchange: definition.Exchange, Timezone: definition.Timezone,
		TradingDayOffset: int64(definition.TradingDayOffset), WeekStart: int64(definition.WeekStart),
		ProductSpecific: definition.ProductSpecific, Remark: definition.Remark,
	}
	for _, session := range definition.Sessions {
		if session == nil {
			continue
		}
		calendar.Sessions = append(calendar.Sessions, &market.MarketTradingSession{
			Id: session.Id, SessionType: session.SessionType, StartTime: session.StartTime,
			EndTime: session.EndTime, CrossDay: session.CrossDay == 1,
			WeekdayMask: session.WeekdayMask, Sort: session.Sort,
		})
	}
	out.TradingCalendar = calendar
	return out
}
