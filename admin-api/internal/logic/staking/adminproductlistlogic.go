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

type AdminProductListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminProductListLogic {
	return &AdminProductListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminProductListLogic) AdminProductList(req *types.AdminProductListReq) (resp *types.AdminProductListResp, err error) {
	return logicutil.Proxy[types.AdminProductListResp](l.ctx, req, l.svcCtx.StakingCli.AdminProductList)
}
