package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminAddAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminAddAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminAddAssetLogic {
	return &AdminAddAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台人工加币
func (l *AdminAddAssetLogic) AdminAddAsset(in *asset.AdminAddAssetReq) (*asset.AdminChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.AdminChangeAssetResp{}, nil
}
