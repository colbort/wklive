package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVisibleCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVisibleCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVisibleCategoriesLogic {
	return &ListVisibleCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取允许显示的产品类型
func (l *ListVisibleCategoriesLogic) ListVisibleCategories(in *itick.ListVisibleCategoriesReq) (*itick.ListVisibleCategoriesResp, error) {
	// todo: add your logic here and delete this line

	return &itick.ListVisibleCategoriesResp{}, nil
}
