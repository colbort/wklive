package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageUserAssetsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageUserAssetsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageUserAssetsLogic {
	return &PageUserAssetsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询资产
func (l *PageUserAssetsLogic) PageUserAssets(in *asset.PageUserAssetsReq) (*asset.PageUserAssetsResp, error) {
	// todo: add your logic here and delete this line

	return &asset.PageUserAssetsResp{}, nil
}
