package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysCronJobDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobDeleteLogic {
	return &SysCronJobDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除系统定时任务
func (l *SysCronJobDeleteLogic) SysCronJobDelete(in *system.SysCronJobDeleteReq) (*system.RespBase, error) {
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
	// 先停止任务
	l.svcCtx.Cron.PauseJob(job.Id)
	// 从调度器中移除任务
	l.svcCtx.Cron.RemoveJob(job.Id)
	// 再删除任务
	err = l.svcCtx.JobModel.Delete(l.ctx, in.Id)
	if err != nil {
		return &system.RespBase{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &system.RespBase{
		Code: 200,
		Msg:  "删除成功",
	}, nil
}
