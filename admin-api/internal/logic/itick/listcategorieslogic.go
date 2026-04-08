// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCategoriesLogic) ListCategories(req *types.ListCategoriesReq) (resp *types.ListCategoriesResp, err error) {
	result, err := l.svcCtx.ItickCli.ListCategories(l.ctx, &itick.ListCategoriesReq{
		Page: &common.PageReq{
			Cursor: req.PageReq.Cursor,
			Limit:  req.PageReq.Limit,
		},
		CategoryType: itick.CategoryType(req.CategoryType),
		Enabled:      req.Enabled,
		AppVisible:   req.AppVisible,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.ItickCategory, 0)
	for _, v := range result.Data {
		data = append(data, types.ItickCategory{
			Id:           v.Id,
			CategoryType: int64(v.CategoryType),
			CategoryCode: v.CategoryCode,
			CategoryName: v.CategoryName,
			Enabled:      v.Enabled,
			AppVisible:   v.AppVisible,
			Sort:         v.Sort,
			Icon:         v.Icon,
			Remark:       v.Remark,
			CreateTimes:  v.CreateTimes,
			UpdateTimes:  v.UpdateTimes,
		})
	}
	return &types.ListCategoriesResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}, nil
}
