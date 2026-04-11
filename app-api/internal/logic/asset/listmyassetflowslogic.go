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

type ListMyAssetFlowsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyAssetFlowsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyAssetFlowsLogic {
	return &ListMyAssetFlowsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyAssetFlowsLogic) ListMyAssetFlows(req *types.ListMyAssetFlowsReq) (resp *types.ListMyAssetFlowsResp, err error) {
	return logicutil.Proxy[types.ListMyAssetFlowsResp](l.ctx, req, l.svcCtx.AssetCli.ListMyAssetFlows)
}
