// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"wklive/admin-api/internal/logicutil"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobUpdateLogic {
	return &SysCronJobUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobUpdateLogic) SysCronJobUpdate(req *types.SysCronJobUpdateReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.SystemCli.SysCronJobUpdate)
}
