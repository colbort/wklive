// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobStopLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobStopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobStopLogic {
	return &SysCronJobStopLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobStopLogic) SysCronJobStop(req *types.SysCronJobStopReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysCronJobStop(l.ctx, &system.SysCronJobStopReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.RespBase{
		Code: result.Code,
		Msg:  result.Msg,
	}
	return
}
