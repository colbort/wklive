package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductCategoriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductCategoriesLogic {
	return &ListProductCategoriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 产品类型列表
func (l *ListProductCategoriesLogic) ListProductCategories(in *itick.ListProductCategoriesReq) (*itick.ListProductCategoriesResp, error) {
	// todo: add your logic here and delete this line

	return &itick.ListProductCategoriesResp{}, nil
}
