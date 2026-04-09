package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminFreezeAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminFreezeAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminFreezeAssetLogic {
	return &AdminFreezeAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台冻结资产
func (l *AdminFreezeAssetLogic) AdminFreezeAsset(in *asset.AdminFreezeAssetReq) (*asset.AdminChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.AdminChangeAssetResp{}, nil
}
