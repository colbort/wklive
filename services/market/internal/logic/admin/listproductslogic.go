package adminlogic

import (
	"context"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 产品列表
func (l *ListProductsLogic) ListProducts(in *market.ListProductsReq) (*market.ListProductsResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	items, count, err := l.svcCtx.MarketProductModel.FindPage(l.ctx, models.MarketProductPageFilter{
		CategoryType: int32(in.CategoryType),
		CategoryName: in.CategoryName,
		Market:       in.Market,
		Keyword:      in.Keyword,
		Enabled:      int32(in.Enabled),
		AppVisible:   int32(in.AppVisible),
		Symbol:       in.Symbol,
	}, cursor, limit)
	if err != nil {
		return nil, err
	}

	hasNext := int64(len(items)) > limit
	if hasNext {
		items = items[:limit]
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	var data []*market.MarketProduct
	for _, item := range items {
		data = append(data, helpers.ToProductProto(item))
	}

	nextCursor := int64(0)
	if hasNext {
		nextCursor = lastID
	}
	return &market.ListProductsResp{
		Base: helper.OkWithOthers(count, hasNext, cursor > 0, nextCursor, cursor),
		Data: data,
	}, nil
}
