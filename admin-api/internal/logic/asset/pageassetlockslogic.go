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

type PageAssetLocksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageAssetLocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetLocksLogic {
	return &PageAssetLocksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageAssetLocksLogic) PageAssetLocks(req *types.PageAssetLocksReq) (resp *types.PageAssetLocksResp, err error) {
	return logicutil.Proxy[types.PageAssetLocksResp](l.ctx, req, l.svcCtx.AssetCli.PageAssetLocks)
}
