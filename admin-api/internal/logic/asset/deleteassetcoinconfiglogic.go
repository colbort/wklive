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

type DeleteAssetCoinConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAssetCoinConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAssetCoinConfigLogic {
	return &DeleteAssetCoinConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAssetCoinConfigLogic) DeleteAssetCoinConfig(req *types.DeleteAssetCoinConfigReq) (resp *types.DeleteAssetCoinConfigResp, err error) {
	return logicutil.Proxy[types.DeleteAssetCoinConfigResp](l.ctx, req, l.svcCtx.AssetCli.DeleteAssetCoinConfig)
}
