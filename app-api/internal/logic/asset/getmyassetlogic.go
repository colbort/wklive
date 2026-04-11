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

type GetMyAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMyAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyAssetLogic {
	return &GetMyAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMyAssetLogic) GetMyAsset(req *types.GetMyAssetReq) (resp *types.GetMyAssetResp, err error) {
	return logicutil.Proxy[types.GetMyAssetResp](l.ctx, req, l.svcCtx.AssetCli.GetMyAsset)
}
