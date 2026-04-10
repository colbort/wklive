package logic

import (
	"context"
	"errors"

	pageutil "wklive/common/pageutil"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListTradesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListTradesLogic {
	return &AppListTradesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取成交记录列表
func (l *AppListTradesLogic) AppListTrades(in *option.AppListTradesReq) (*option.AppListTradesResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionTradeModel.FindPage(l.ctx, models.OptionTradePageFilter{
		TenantId:       in.TenantId,
		ContractId:     in.ContractId,
		Uid:            in.Uid,
		AccountId:      in.AccountId,
		TradeTimeStart: pageutil.TimeRangeStart(in.TradeTimeRange),
		TradeTimeEnd:   pageutil.TimeRangeEnd(in.TradeTimeRange),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	list := make([]*option.OptionTradeDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, err := buildTradeDetail(l.ctx, l.svcCtx, item)
		if err != nil {
			return nil, err
		}
		list = append(list, detail)
	}

	return &option.AppListTradesResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}
