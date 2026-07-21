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

type SetPlatformAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetPlatformAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetPlatformAccountLogic {
	return &SetPlatformAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetPlatformAccountLogic) SetPlatformAccount(req *types.SetPlatformAccountReq) (resp *types.PlatformAccountResp, err error) {
	return logicutil.Proxy[types.PlatformAccountResp](l.ctx, req, l.svcCtx.AssetCli.SetPlatformAccount)
}
