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

type PageAssetCoinConfigsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageAssetCoinConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetCoinConfigsLogic {
	return &PageAssetCoinConfigsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageAssetCoinConfigsLogic) PageAssetCoinConfigs(req *types.PageAssetCoinConfigsReq) (resp *types.PageAssetCoinConfigsResp, err error) {
	return logicutil.Proxy[types.PageAssetCoinConfigsResp](l.ctx, req, l.svcCtx.AssetCli.PageAssetCoinConfigs)
}
