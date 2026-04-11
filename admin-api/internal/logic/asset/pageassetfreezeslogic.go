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

type PageAssetFreezesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageAssetFreezesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetFreezesLogic {
	return &PageAssetFreezesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageAssetFreezesLogic) PageAssetFreezes(req *types.PageAssetFreezesReq) (resp *types.PageAssetFreezesResp, err error) {
	return logicutil.Proxy[types.PageAssetFreezesResp](l.ctx, req, l.svcCtx.AssetCli.PageAssetFreezes)
}
