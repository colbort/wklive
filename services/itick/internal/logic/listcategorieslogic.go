package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

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
func (l *ListCategoriesLogic) ListCategories(in *itick.ListCategoriesReq) (*itick.ListCategoriesResp, error) {
	items, count, err := l.svcCtx.ItickCategoryModel.FindPage(l.ctx, int32(in.CategoryType), in.Enabled, in.AppVisible, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	var data []*itick.ItickCategory
	for _, item := range items {
		data = append(data, &itick.ItickCategory{
			Id:           item.Id,
			CategoryType: itick.CategoryType(item.CategoryType),
			CategoryCode: item.CategoryCode,
			CategoryName: item.CategoryName,
			Enabled:      item.Enabled,
			AppVisible:   item.AppVisible,
			Sort:         item.Sort,
			Icon:         item.Icon,
			Remark:       item.Remark,
			CreateTimes:  item.CreateTimes,
			UpdateTimes:  item.UpdateTimes,
		})
	}

	return &itick.ListCategoriesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), count, lastID),
		Data: data,
	}, nil
}
