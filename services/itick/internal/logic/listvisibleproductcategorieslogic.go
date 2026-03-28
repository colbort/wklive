package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListVisibleProductCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListVisibleProductCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListVisibleProductCategoriesLogic {
	return &ListVisibleProductCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取允许显示的产品类型
func (l *ListVisibleProductCategoriesLogic) ListVisibleProductCategories(in *itick.ListVisibleProductCategoriesReq) (*itick.ListVisibleProductCategoriesResp, error) {
	// todo: add your logic here and delete this line

	return &itick.ListVisibleProductCategoriesResp{}, nil
}
