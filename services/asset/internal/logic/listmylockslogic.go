package logic

import (
	"context"

	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyLocksLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyLocksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyLocksLogic {
	return &ListMyLocksLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的锁仓明细
func (l *ListMyLocksLogic) ListMyLocks(in *asset.ListMyLocksReq) (*asset.ListMyLocksResp, error) {
	// todo: add your logic here and delete this line

	return &asset.ListMyLocksResp{}, nil
}
