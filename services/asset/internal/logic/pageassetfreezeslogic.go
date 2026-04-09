package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageAssetFreezesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageAssetFreezesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetFreezesLogic {
	return &PageAssetFreezesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询冻结明细
func (l *PageAssetFreezesLogic) PageAssetFreezes(in *asset.PageAssetFreezesReq) (*asset.PageAssetFreezesResp, error) {
	// todo: add your logic here and delete this line

	return &asset.PageAssetFreezesResp{}, nil
}
