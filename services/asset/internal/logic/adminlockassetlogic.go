package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminLockAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminLockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLockAssetLogic {
	return &AdminLockAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台锁仓资产
func (l *AdminLockAssetLogic) AdminLockAsset(in *asset.AdminLockAssetReq) (*asset.AdminChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.AdminChangeAssetResp{}, nil
}
