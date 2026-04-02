// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVisibleCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListVisibleCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVisibleCategoriesLogic {
	return &ListVisibleCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListVisibleCategoriesLogic) ListVisibleCategories(req *types.ListVisibleCategoriesReq) (resp *types.ListVisibleCategoriesResp, err error) {
	result, err := l.svcCtx.ItickCli.ListVisibleCategories(l.ctx, &itick.ListVisibleCategoriesReq{})
	if err != nil {
		return nil, err
	}

	resp = &types.ListVisibleCategoriesResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: make([]types.ItickTenantCategory, 0, len(result.Data)),
	}
	for _, item := range result.Data {
		resp.Data = append(resp.Data, types.ItickTenantCategory{
			Id:             item.Id,
			TenantId:       item.TenantId,
			CategoryId:     item.CategoryId,
			Enabled:        item.Enabled,
			AppVisible:     item.AppVisible,
			Sort:           item.Sort,
			Remark:         item.Remark,
			CreateTime:     item.CreateTime,
			UpdateTime:     item.UpdateTime,
			CategoryType:   int64(item.CategoryType),
			CategoryName:   item.CategoryName,
			Icon:           item.Icon,
		})
	}

	return
}
