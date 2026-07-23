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

type UnfreezeAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnfreezeAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfreezeAssetLogic {
	return &UnfreezeAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnfreezeAssetLogic) UnfreezeAsset(req *types.UnfreezeAssetReq) (resp *types.ChangeAssetResp, err error) {
	return logicutil.Proxy[types.ChangeAssetResp](l.ctx, req, l.svcCtx.AssetCli.UnfreezeAsset)
}
