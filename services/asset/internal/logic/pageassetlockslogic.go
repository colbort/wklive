package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageAssetLocksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageAssetLocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetLocksLogic {
	return &PageAssetLocksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询锁仓明细
func (l *PageAssetLocksLogic) PageAssetLocks(in *asset.PageAssetLocksReq) (*asset.PageAssetLocksResp, error) {
	// todo: add your logic here and delete this line

	return &asset.PageAssetLocksResp{}, nil
}
