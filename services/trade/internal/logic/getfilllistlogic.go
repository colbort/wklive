package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFillListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFillListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFillListLogic {
	return &GetFillListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取成交记录列表
func (l *GetFillListLogic) GetFillList(in *trade.GetFillListReq) (*trade.GetFillListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	list, total, err := l.svcCtx.TradeFillModel.FindPage(l.ctx, models.TradeFillPageFilter{
		TenantId:   int64(in.TenantId),
		UserId:     int64(in.UserId),
		SymbolId:   int64(in.SymbolId),
		MarketType: int64(in.MarketType),
		TimeStart:  in.TimeRange.StartTime,
		TimeEnd:    in.TimeRange.EndTime,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(list) > 0 {
		lastID = int64(list[len(list)-1].Id)
	}
	resp := &trade.GetFillListResp{Base: pageutil.Base(cursor, limit, len(list), total, lastID)}
	for _, item := range list {
		resp.List = append(resp.List, fillToProto(item))
	}
	return resp, nil
}
