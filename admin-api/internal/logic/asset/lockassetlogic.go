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

type LockAssetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockAssetLogic {
	return &LockAssetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LockAssetLogic) LockAsset(req *types.LockAssetReq) (resp *types.ChangeAssetResp, err error) {
	return logicutil.Proxy[types.ChangeAssetResp](l.ctx, req, l.svcCtx.AssetCli.LockAsset)
}
