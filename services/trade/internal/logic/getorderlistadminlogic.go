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

type GetOrderListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderListAdminLogic {
	return &GetOrderListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取后台订单列表
func (l *GetOrderListAdminLogic) GetOrderListAdmin(in *trade.GetOrderListAdminReq) (*trade.GetOrderListAdminResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	list, total, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
		TenantId:   int64(in.TenantId),
		UserId:     int64(in.UserId),
		SymbolId:   int64(in.SymbolId),
		MarketType: int64(in.MarketType),
		Status:     int64(in.Status),
		TimeStart:  in.TimeRange.StartTime,
		TimeEnd:    in.TimeRange.EndTime,
		Keyword:    in.Keyword,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(list) > 0 {
		lastID = int64(list[len(list)-1].Id)
	}
	resp := &trade.GetOrderListAdminResp{Base: pageutil.Base(cursor, limit, len(list), total, lastID)}
	for _, item := range list {
		resp.List = append(resp.List, orderToProto(item))
	}
	return resp, nil
}
