package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
)

type SysCronJobStopLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysCronJobStopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobStopLogic {
	return &SysCronJobStopLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 停止系统定时任务
func (l *SysCronJobStopLogic) SysCronJobStop(in *system.SysCronJobStopReq) (*system.RespBase, error) {
	job, err := l.svcCtx.JobModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}
	if job == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(400, i18n.Translate(i18n.CronJobNotFound, l.ctx)),
		}, nil
	}
	l.svcCtx.Cron.PauseJob(job.Id)
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
