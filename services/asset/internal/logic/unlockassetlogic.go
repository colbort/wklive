package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlockAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlockAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockAssetLogic {
	return &UnlockAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 解锁
func (l *UnlockAssetLogic) UnlockAsset(in *asset.UnlockAssetReq) (*asset.ChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.ChangeAssetResp{}, nil
}
