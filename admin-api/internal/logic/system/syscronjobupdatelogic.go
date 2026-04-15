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
	result, err := l.svcCtx.SystemCli.SysCronJobUpdate(l.ctx, &system.SysCronJobUpdateReq{
		Id:             req.Id,
		JobName:        req.JobName,
		JobGroup:       req.JobGroup,
		InvokeTarget:   req.InvokeTarget,
		CronExpression: req.CronExpression,
		Status:         toJobStatus(req.Status),
		Remark:         req.Remark,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}
	return
}
