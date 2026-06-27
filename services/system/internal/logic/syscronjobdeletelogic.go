package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
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
	if base, err := systemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}

	job, err := l.svcCtx.JobModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.InternalServerError, i18n.Translate(i18n.InternalServerError, l.ctx)),
		}, nil
	}
	if job == nil {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.CronJobNotFound, i18n.Translate(i18n.CronJobNotFound, l.ctx)),
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
			Base: helper.ErrResp(i18n.InternalServerError, i18n.Translate(i18n.InternalServerError, l.ctx)),
		}, nil
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
