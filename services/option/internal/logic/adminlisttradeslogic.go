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

type AdminListTradesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListTradesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListTradesLogic {
	return &AdminListTradesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询成交记录列表
func (l *AdminListTradesLogic) AdminListTrades(in *option.ListTradesReq) (*option.ListTradesResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionTradeModel.FindPage(l.ctx, models.OptionTradePageFilter{
		TenantId:       in.TenantId,
		ContractId:     in.ContractId,
		Uid:            in.Uid,
		TradeNo:        in.TradeNo,
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

	return &option.ListTradesResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}
