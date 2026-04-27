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

type UpdateAssetCoinConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateAssetCoinConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAssetCoinConfigLogic {
	return &UpdateAssetCoinConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAssetCoinConfigLogic) UpdateAssetCoinConfig(req *types.UpdateAssetCoinConfigReq) (resp *types.AssetCoinConfigResp, err error) {
	return logicutil.Proxy[types.AssetCoinConfigResp](l.ctx, req, l.svcCtx.AssetCli.UpdateAssetCoinConfig)
}
