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

type PageAssetFlowsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageAssetFlowsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetFlowsLogic {
	return &PageAssetFlowsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageAssetFlowsLogic) PageAssetFlows(req *types.PageAssetFlowsReq) (resp *types.PageAssetFlowsResp, err error) {
	return logicutil.Proxy[types.PageAssetFlowsResp](l.ctx, req, l.svcCtx.AssetCli.PageAssetFlows)
}
