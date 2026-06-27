package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
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
			Base: helper.ErrResp(i18n.InternalServerError, i18n.Translate(i18n.InternalServerError, l.ctx)),
		}, nil
	}
	if job == nil {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.CronJobNotFound, i18n.Translate(i18n.CronJobNotFound, l.ctx)),
		}, nil
	}
	err = l.svcCtx.Cron.RunOnce(job)
	if err != nil {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.InternalServerError, i18n.Translate(i18n.InternalServerError, l.ctx)),
		}, nil
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
