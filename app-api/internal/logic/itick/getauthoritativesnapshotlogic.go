// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package itick

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuthoritativeSnapshotLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAuthoritativeSnapshotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuthoritativeSnapshotLogic {
	return &GetAuthoritativeSnapshotLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAuthoritativeSnapshotLogic) GetAuthoritativeSnapshot(req *types.GetAuthoritativeSnapshotReq) (resp *types.GetAuthoritativeSnapshotResp, err error) {
	return logicutil.Proxy[types.GetAuthoritativeSnapshotResp](l.ctx, req, l.svcCtx.ItickCli.GetAuthoritativeSnapshot)
}
