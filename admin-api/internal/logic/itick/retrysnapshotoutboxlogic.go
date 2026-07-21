// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package itick

import (
	"context"
	"wklive/admin-api/internal/logicutil"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetrySnapshotOutboxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetrySnapshotOutboxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetrySnapshotOutboxLogic {
	return &RetrySnapshotOutboxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetrySnapshotOutboxLogic) RetrySnapshotOutbox(req *types.RetrySnapshotOutboxReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.ItickCli.RetrySnapshotOutbox)
}
