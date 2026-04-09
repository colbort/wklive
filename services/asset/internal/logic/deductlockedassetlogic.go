package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductLockedAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductLockedAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductLockedAssetLogic {
	return &DeductLockedAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 扣减锁仓余额
func (l *DeductLockedAssetLogic) DeductLockedAsset(in *asset.DeductLockedAssetReq) (*asset.ChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.ChangeAssetResp{}, nil
}
