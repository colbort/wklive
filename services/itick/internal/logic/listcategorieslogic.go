package logic

import (
	"context"

	"wklive/proto/common"
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

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit

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
		Base: &common.RespBase{
			Code:       200,
			Msg:        "查询成功",
			Total:      count,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
		},
		Data: data,
	}, nil
}
