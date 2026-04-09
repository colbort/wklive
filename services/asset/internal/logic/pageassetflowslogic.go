package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageAssetFlowsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageAssetFlowsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetFlowsLogic {
	return &PageAssetFlowsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询资产流水
func (l *PageAssetFlowsLogic) PageAssetFlows(in *asset.PageAssetFlowsReq) (*asset.PageAssetFlowsResp, error) {
	// todo: add your logic here and delete this line

	return &asset.PageAssetFlowsResp{}, nil
}
