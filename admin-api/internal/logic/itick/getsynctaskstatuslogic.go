// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSyncTaskStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSyncTaskStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSyncTaskStatusLogic {
	return &GetSyncTaskStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSyncTaskStatusLogic) GetSyncTaskStatus(req *types.GetSyncTaskStatusReq) (resp *types.GetSyncTaskStatusResp, err error) {
	return logicutil.Proxy[types.GetSyncTaskStatusResp](l.ctx, req, l.svcCtx.ItickCli.GetSyncTaskStatus)
}
