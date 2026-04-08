package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductCreateLogic {
	return &ProductCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建质押产品
func (l *ProductCreateLogic) ProductCreate(in *staking.AdminProductCreateReq) (*staking.AdminProductCreateResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AdminProductCreateResp{}, nil
}
