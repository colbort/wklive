package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/common/utils"
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
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.BizTradeEventModel.FindPage(l.ctx, models.BizTradeEventPageFilter{
		TenantId:    in.TenantId,
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
	if len(data) > 0 {
		lastID = int64(data[len(data)-1].Id)
	}
	resp := &trade.GetTradeEventListResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		resp.Data = append(resp.Data, tradeEventToProto(item))
	}
	return resp, nil
}
