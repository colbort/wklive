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
	result, err := l.svcCtx.SystemCli.SysCronJobCreate(l.ctx, &system.SysCronJobCreateReq{
		JobName:        req.JobName,
		JobGroup:       req.JobGroup,
		InvokeTarget:   req.InvokeTarget,
		CronExpression: req.CronExpression,
		Status:         req.Status,
		Remark:         req.Remark,
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
