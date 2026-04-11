// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package asset

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyLocksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyLocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyLocksLogic {
	return &ListMyLocksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyLocksLogic) ListMyLocks(req *types.ListMyLocksReq) (resp *types.ListMyLocksResp, err error) {
	return logicutil.Proxy[types.ListMyLocksResp](l.ctx, req, l.svcCtx.AssetCli.ListMyLocks)
}
