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

type GetTradeEventListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTradeEventListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTradeEventListLogic {
	return &GetTradeEventListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取交易事件列表
func (l *GetTradeEventListLogic) GetTradeEventList(in *trade.GetTradeEventListReq) (*trade.GetTradeEventListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	list, total, err := l.svcCtx.BizTradeEventModel.FindPage(l.ctx, models.BizTradeEventPageFilter{
		TenantId:    int64(in.TenantId),
		EventType:   in.EventType,
		BizType:     in.BizType,
		BizId:       in.BizId,
		EventStatus: int64(in.EventStatus),
		TimeStart:   in.TimeRange.StartTime,
		TimeEnd:     in.TimeRange.EndTime,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	lastID := int64(0)
	if len(list) > 0 {
		lastID = int64(list[len(list)-1].Id)
	}
	resp := &trade.GetTradeEventListResp{Base: pageutil.Base(cursor, limit, len(list), total, lastID)}
	for _, item := range list {
		resp.List = append(resp.List, tradeEventToProto(item))
	}
	return resp, nil
}
