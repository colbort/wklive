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

type GetPlatformAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPlatformAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlatformAccountLogic {
	return &GetPlatformAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPlatformAccountLogic) GetPlatformAccount(req *types.GetPlatformAccountReq) (resp *types.PlatformAccountResp, err error) {
	return logicutil.Proxy[types.PlatformAccountResp](l.ctx, req, l.svcCtx.AssetCli.GetPlatformAccount)
}
