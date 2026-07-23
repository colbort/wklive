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

type AddAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAssetLogic {
	return &AddAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddAssetLogic) AddAsset(req *types.AddAssetReq) (resp *types.ChangeAssetResp, err error) {
	return logicutil.Proxy[types.ChangeAssetResp](l.ctx, req, l.svcCtx.AssetCli.AddAsset)
}
