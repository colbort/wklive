package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProductCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductCategoryLogic {
	return &GetProductCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取产品类型详情
func (l *GetProductCategoryLogic) GetProductCategory(in *itick.GetProductCategoryReq) (*itick.GetProductCategoryResp, error) {
	// todo: add your logic here and delete this line

	return &itick.GetProductCategoryResp{}, nil
}
