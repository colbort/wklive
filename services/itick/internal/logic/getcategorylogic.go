package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCategoryLogic {
	return &GetCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取产品类型详情
func (l *GetCategoryLogic) GetCategory(in *itick.GetCategoryReq) (*itick.GetCategoryResp, error) {
	// todo: add your logic here and delete this line

	return &itick.GetCategoryResp{}, nil
}
