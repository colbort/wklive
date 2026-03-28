package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新产品仅允许更新名称、状态、排序、图标和备注，市场、品种、代码不允许修改
func (l *UpdateProductLogic) UpdateProduct(in *itick.UpdateProductReq) (*itick.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &itick.AdminCommonResp{}, nil
}
