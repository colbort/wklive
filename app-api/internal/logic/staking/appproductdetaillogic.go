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

type AppProductDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppProductDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppProductDetailLogic {
	return &AppProductDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppProductDetailLogic) AppProductDetail(req *types.AppProductDetailReq) (resp *types.AppProductDetailResp, err error) {
	return logicutil.Proxy[types.AppProductDetailResp](l.ctx, req, l.svcCtx.StakingCli.AppProductDetail)
}
