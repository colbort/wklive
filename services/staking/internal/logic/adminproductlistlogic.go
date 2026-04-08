package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminProductListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminProductListLogic {
	return &AdminProductListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押产品列表
func (l *AdminProductListLogic) AdminProductList(in *staking.AdminProductListReq) (*staking.AdminProductListResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AdminProductListResp{}, nil
}
