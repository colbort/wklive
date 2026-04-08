package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductChangeStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductChangeStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductChangeStatusLogic {
	return &ProductChangeStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改质押产品状态
func (l *ProductChangeStatusLogic) ProductChangeStatus(in *staking.AdminProductChangeStatusReq) (*staking.AdminProductChangeStatusResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AdminProductChangeStatusResp{}, nil
}
