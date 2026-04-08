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

type ListTenantCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTenantCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantCategoriesLogic {
	return &ListTenantCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTenantCategoriesLogic) ListTenantCategories(req *types.ListTenantCategoriesReq) (resp *types.ListTenantCategoriesResp, err error) {
	result, err := l.svcCtx.ItickCli.ListTenantCategories(l.ctx, &itick.ListTenantCategoriesReq{
		Page: &common.PageReq{
			Cursor: req.PageReq.Cursor,
			Limit:  req.PageReq.Limit,
		},
		TenantId:      req.TenantId,
		CategoryType:  itick.CategoryType(req.CategoryType),
		Status:        req.Status,
		VisibleStatus: req.VisibleStatus,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.ItickTenantCategory, 0)
	for _, item := range result.Data {
		data = append(data, types.ItickTenantCategory{
			Id:           item.Id,
			TenantId:     item.TenantId,
			CategoryId:   item.CategoryId,
			Enabled:      item.Enabled,
			AppVisible:   item.AppVisible,
			Sort:         item.Sort,
			Remark:       item.Remark,
			CreateTimes:  item.CreateTimes,
			UpdateTimes:  item.UpdateTimes,
			CategoryType: int64(item.CategoryType),
			CategoryName: item.CategoryName,
			Icon:         item.Icon,
		})
	}

	return &types.ListTenantCategoriesResp{
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
