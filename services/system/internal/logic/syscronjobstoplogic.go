package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}
	if job == nil {
		return &system.RespBase{
			Code: 400,
			Msg:  "定时任务不存在",
		}, nil
	}
	l.svcCtx.Cron.PauseJob(job.Id)
	return &system.RespBase{}, nil
}
