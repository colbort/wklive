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

type ListMyAssetsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyAssetsLogic {
	return &ListMyAssetsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyAssetsLogic) ListMyAssets(req *types.ListMyAssetsReq) (resp *types.ListMyAssetsResp, err error) {
	return logicutil.Proxy[types.ListMyAssetsResp](l.ctx, req, l.svcCtx.AssetCli.ListMyAssets)
}
