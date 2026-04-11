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

type GetMyAssetSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyAssetSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyAssetSummaryLogic {
	return &GetMyAssetSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyAssetSummaryLogic) GetMyAssetSummary(req *types.GetMyAssetSummaryReq) (resp *types.GetMyAssetSummaryResp, err error) {
	return logicutil.Proxy[types.GetMyAssetSummaryResp](l.ctx, req, l.svcCtx.AssetCli.GetMyAssetSummary)
}
