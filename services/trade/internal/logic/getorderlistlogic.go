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

type GetOrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderListLogic {
	return &GetOrderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取订单列表
func (l *GetOrderListLogic) GetOrderList(in *trade.GetOrderListReq) (*trade.GetOrderListResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
		TenantId:   tenantId,
		UserId:     userId,
		SymbolId:   in.SymbolId,
		MarketType: int64(in.MarketType),
		Status:     int64(in.Status),
		Side:       int64(in.Side),
		TimeStart:  in.TimeRange.StartTime,
		TimeEnd:    in.TimeRange.EndTime,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = int64(data[len(data)-1].Id)
	}
	resp := &trade.GetOrderListResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		resp.Data = append(resp.Data, orderToProto(item))
	}
	return resp, nil
}
