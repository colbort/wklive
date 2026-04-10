package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobRunLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysCronJobRunLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobRunLogic {
	return &SysCronJobRunLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 执行一次系统定时任务
func (l *SysCronJobRunLogic) SysCronJobRun(in *system.SysCronJobRunReq) (*system.RespBase, error) {
	job, err := l.svcCtx.JobModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}
	if job == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(400, "定时任务不存在"),
		}, nil
	}
	err = l.svcCtx.Cron.RunOnce(job)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
