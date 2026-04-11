package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolListLogic {
	return &GetSymbolListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取交易对列表
func (l *GetSymbolListLogic) GetSymbolList(in *trade.GetSymbolListReq) (*trade.GetSymbolListResp, error) {
	list, err := l.svcCtx.TradeSymbolModel.FindAll(l.ctx, models.TradeSymbolPageFilter{
		TenantId:   int64(in.TenantId),
		MarketType: int64(in.MarketType),
		Status:     int64(in.Status),
	})
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	resp := &trade.GetSymbolListResp{Base: helper.OkResp()}
	for _, item := range list {
		resp.List = append(resp.List, symbolToProto(item))
	}
	return resp, nil
}
