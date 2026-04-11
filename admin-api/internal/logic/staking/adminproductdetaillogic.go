// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminProductDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminProductDetailLogic {
	return &AdminProductDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminProductDetailLogic) AdminProductDetail(req *types.AdminProductDetailReq) (resp *types.AdminProductDetailResp, err error) {
	return logicutil.Proxy[types.AdminProductDetailResp](l.ctx, req, l.svcCtx.StakingCli.AdminProductDetail)
}
