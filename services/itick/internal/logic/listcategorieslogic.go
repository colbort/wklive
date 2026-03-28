package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &itick.ListCategoriesResp{}, nil
}
