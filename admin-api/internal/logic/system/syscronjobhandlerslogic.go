// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobHandlersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobHandlersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobHandlersLogic {
	return &SysCronJobHandlersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobHandlersLogic) SysCronJobHandlers() (resp *types.SysCronJobHandlersResp, err error) {
	return logicutil.Proxy[types.SysCronJobHandlersResp](l.ctx, nil, l.svcCtx.SystemCli.SysCronJobHandlers)
}
