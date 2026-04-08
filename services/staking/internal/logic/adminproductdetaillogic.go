package logic

import (
	"context"

	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminProductDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminProductDetailLogic {
	return &AdminProductDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取质押产品详情
func (l *AdminProductDetailLogic) AdminProductDetail(in *staking.AdminProductDetailReq) (*staking.AdminProductDetailResp, error) {
	// todo: add your logic here and delete this line

	return &staking.AdminProductDetailResp{}, nil
}
