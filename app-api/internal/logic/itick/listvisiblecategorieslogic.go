// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
