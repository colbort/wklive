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

type AdminUnfreezeAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUnfreezeAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUnfreezeAssetLogic {
	return &AdminUnfreezeAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUnfreezeAssetLogic) AdminUnfreezeAsset(req *types.AdminUnfreezeAssetReq) (resp *types.AdminChangeAssetResp, err error) {
	return logicutil.Proxy[types.AdminChangeAssetResp](l.ctx, req, l.svcCtx.AssetCli.AdminUnfreezeAsset)
}
