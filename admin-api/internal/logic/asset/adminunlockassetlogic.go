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

type AdminUnlockAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUnlockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUnlockAssetLogic {
	return &AdminUnlockAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUnlockAssetLogic) AdminUnlockAsset(req *types.AdminUnlockAssetReq) (resp *types.AdminChangeAssetResp, err error) {
	return logicutil.Proxy[types.AdminChangeAssetResp](l.ctx, req, l.svcCtx.AssetCli.AdminUnlockAsset)
}
