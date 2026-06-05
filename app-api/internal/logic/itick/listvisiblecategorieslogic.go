// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/common"
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
	return logicutil.Proxy[types.ListVisibleCategoriesResp](l.ctx, &itick.ListVisibleCategoriesReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
	}, l.svcCtx.ItickCli.ListVisibleCategories)
}
