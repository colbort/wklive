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

type AdminFreezeAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminFreezeAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminFreezeAssetLogic {
	return &AdminFreezeAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminFreezeAssetLogic) AdminFreezeAsset(req *types.AdminFreezeAssetReq) (resp *types.AdminChangeAssetResp, err error) {
	return logicutil.Proxy[types.AdminChangeAssetResp](l.ctx, req, l.svcCtx.AssetCli.AdminFreezeAsset)
}
