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

type AdminSubAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminSubAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSubAssetLogic {
	return &AdminSubAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminSubAssetLogic) AdminSubAsset(req *types.AdminSubAssetReq) (resp *types.AdminChangeAssetResp, err error) {
	return logicutil.Proxy[types.AdminChangeAssetResp](l.ctx, req, l.svcCtx.AssetCli.AdminSubAsset)
}
