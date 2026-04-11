// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppProductListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppProductListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppProductListLogic {
	return &AppProductListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppProductListLogic) AppProductList(req *types.AppProductListReq) (resp *types.AppProductListResp, err error) {
	return logicutil.Proxy[types.AppProductListResp](l.ctx, req, l.svcCtx.StakingCli.AppProductList)
}
