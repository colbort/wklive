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

type SysCronJobListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobListLogic {
	return &SysCronJobListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobListLogic) SysCronJobList(req *types.SysCronJobListReq) (resp *types.SysCronJobListResp, err error) {
	return logicutil.Proxy[types.SysCronJobListResp](l.ctx, req, l.svcCtx.SystemCli.SysCronJobList)
}
