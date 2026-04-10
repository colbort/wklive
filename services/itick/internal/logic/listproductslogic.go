package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

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
func (l *ListProductsLogic) ListProducts(in *itick.ListProductsReq) (*itick.ListProductsResp, error) {
	items, count, err := l.svcCtx.ItickProductModel.FindPage(l.ctx, int32(in.CategoryType), in.CategoryName, in.Market, in.Keyword, in.Enabled, in.AppVisible, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	var data []*itick.ItickProduct
	for _, item := range items {
		data = append(data, &itick.ItickProduct{
			Id:           item.Id,
			CategoryType: itick.CategoryType(item.CategoryType),
			CategoryName: item.CategoryName,
			CategoryCode: item.CategoryCode,
			Market:       item.Market,
			Symbol:       item.Symbol,
			Code:         item.Code,
			Name:         item.Name,
			DisplayName:  item.DisplayName,
			BaseCoin:     item.BaseCoin,
			QuoteCoin:    item.QuoteCoin,
			Enabled:      item.Enabled,
			AppVisible:   item.AppVisible,
			Sort:         item.Sort,
			Icon:         item.Icon,
			Remark:       item.Remark,
			CreateTimes:  item.CreateTimes,
			UpdateTimes:  item.UpdateTimes,
		})
	}

	return &itick.ListProductsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), count, lastID),
		Data: data,
	}, nil
}
