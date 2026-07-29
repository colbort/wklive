package adminlogic

import (
	"context"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/pageutil"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 产品类型列表
func (l *ListCategoriesLogic) ListCategories(in *market.ListCategoriesReq) (*market.ListCategoriesResp, error) {
	items, count, err := l.svcCtx.MarketCategoryModel.FindPage(l.ctx, models.MarketCategoryPageFilter{
		CategoryType: int32(in.CategoryType),
		Enabled:      int32(in.Enabled),
		AppVisible:   int32(in.AppVisible),
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	var data []*market.MarketCategory
	for _, item := range items {
		data = append(data, helpers.ToCategoryProto(item))
	}

	return &market.ListCategoriesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), count, lastID),
		Data: data,
	}, nil
}
