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

type GetSymbolListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolListAdminLogic {
	return &GetSymbolListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取后台交易对列表
func (l *GetSymbolListAdminLogic) GetSymbolListAdmin(in *trade.GetSymbolListAdminReq) (*trade.GetSymbolListAdminResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.TradeSymbolModel.FindPage(l.ctx, models.TradeSymbolPageFilter{
		TenantId:   in.TenantId,
		MarketType: int64(in.MarketType),
		Status:     int64(in.Status),
		Keyword:    in.Keyword,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = int64(data[len(data)-1].Id)
	}
	resp := &trade.GetSymbolListAdminResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		resp.Data = append(resp.Data, symbolToProto(item))
	}
	return resp, nil
}
