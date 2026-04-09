package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductFrozenAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductFrozenAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductFrozenAssetLogic {
	return &DeductFrozenAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 扣减冻结余额
func (l *DeductFrozenAssetLogic) DeductFrozenAsset(in *asset.DeductFrozenAssetReq) (*asset.ChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.ChangeAssetResp{}, nil
}
