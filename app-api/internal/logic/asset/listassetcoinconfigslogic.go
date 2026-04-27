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

type ListAssetCoinConfigsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAssetCoinConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAssetCoinConfigsLogic {
	return &ListAssetCoinConfigsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAssetCoinConfigsLogic) ListAssetCoinConfigs(req *types.ListAssetCoinConfigsReq) (resp *types.ListAssetCoinConfigsResp, err error) {
	return logicutil.Proxy[types.ListAssetCoinConfigsResp](l.ctx, req, l.svcCtx.AssetCli.ListAssetCoinConfigs)
}
