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

type RevokeAuthoritativeSnapshotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRevokeAuthoritativeSnapshotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeAuthoritativeSnapshotLogic {
	return &RevokeAuthoritativeSnapshotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeAuthoritativeSnapshotLogic) RevokeAuthoritativeSnapshot(req *types.RevokeAuthoritativeSnapshotReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.ItickCli.RevokeAuthoritativeSnapshot)
}
