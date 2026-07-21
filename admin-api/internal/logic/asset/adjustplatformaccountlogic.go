// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package asset

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdjustPlatformAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdjustPlatformAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdjustPlatformAccountLogic {
	return &AdjustPlatformAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdjustPlatformAccountLogic) AdjustPlatformAccount(req *types.AdjustPlatformAccountReq) (resp *types.PlatformAccountResp, err error) {
	return logicutil.Proxy[types.PlatformAccountResp](l.ctx, req, l.svcCtx.AssetCli.AdjustPlatformAccount)
}
