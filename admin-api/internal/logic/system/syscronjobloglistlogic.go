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

type SysCronJobLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobLogListLogic {
	return &SysCronJobLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobLogListLogic) SysCronJobLogList(req *types.SysCronJobLogListReq) (resp *types.SysCronJobLogListResp, err error) {
	return logicutil.Proxy[types.SysCronJobLogListResp](l.ctx, req, l.svcCtx.SystemCli.SysCronJobLogList)
}
