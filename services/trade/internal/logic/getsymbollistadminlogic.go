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
	list, total, err := l.svcCtx.TradeSymbolModel.FindPage(l.ctx, models.TradeSymbolPageFilter{
		TenantId:   int64(in.TenantId),
		MarketType: int64(in.MarketType),
		Status:     int64(in.Status),
		Keyword:    in.Keyword,
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(list) > 0 {
		lastID = int64(list[len(list)-1].Id)
	}
	resp := &trade.GetSymbolListAdminResp{Base: pageutil.Base(cursor, limit, len(list), total, lastID)}
	for _, item := range list {
		resp.List = append(resp.List, symbolToProto(item))
	}
	return resp, nil
}
