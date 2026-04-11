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

type PageUserAssetsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageUserAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageUserAssetsLogic {
	return &PageUserAssetsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageUserAssetsLogic) PageUserAssets(req *types.PageUserAssetsReq) (resp *types.PageUserAssetsResp, err error) {
	return logicutil.Proxy[types.PageUserAssetsResp](l.ctx, req, l.svcCtx.AssetCli.PageUserAssets)
}
