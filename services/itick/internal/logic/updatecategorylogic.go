package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCategoryLogic {
	return &UpdateCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新产品类型仅允许更新名称、状态、排序、图标和备注，产品类型不允许修改
func (l *UpdateCategoryLogic) UpdateCategory(in *itick.UpdateCategoryReq) (*itick.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &itick.AdminCommonResp{}, nil
}
