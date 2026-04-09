package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminSubAssetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminSubAssetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminSubAssetLogic {
	return &AdminSubAssetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 后台人工减币
func (l *AdminSubAssetLogic) AdminSubAsset(in *asset.AdminSubAssetReq) (*asset.AdminChangeAssetResp, error) {
	// todo: add your logic here and delete this line

	return &asset.AdminChangeAssetResp{}, nil
}
