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

type SysCronJobCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobCreateLogic {
	return &SysCronJobCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobCreateLogic) SysCronJobCreate(req *types.SysCronJobCreateReq) (resp *types.RespBase, err error) {
	return logicutil.Proxy[types.RespBase](l.ctx, req, l.svcCtx.SystemCli.SysCronJobCreate)
}
