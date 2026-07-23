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

type SubAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubAssetLogic {
	return &SubAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubAssetLogic) SubAsset(req *types.SubAssetReq) (resp *types.ChangeAssetResp, err error) {
	return logicutil.Proxy[types.ChangeAssetResp](l.ctx, req, l.svcCtx.AssetCli.SubAsset)
}
