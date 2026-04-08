package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductUpdateLogic {
	return &ProductUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新质押产品
func (l *ProductUpdateLogic) ProductUpdate(in *staking.AdminProductUpdateReq) (*staking.AdminProductUpdateResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AdminProductUpdateResp{}, nil
}
