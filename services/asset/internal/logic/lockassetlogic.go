package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LockAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LockAssetLogic {
	return &LockAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 锁仓
func (l *LockAssetLogic) LockAsset(in *asset.LockAssetReq) (*asset.LockAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.LockAssetResp{}, nil
}
