package logic

import (
	"context"
	"database/sql"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysCronJobUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobUpdateLogic {
	return &SysCronJobUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新系统定时任务
func (l *SysCronJobUpdateLogic) SysCronJobUpdate(in *system.SysCronJobUpdateReq) (*system.RespBase, error) {
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	if in.CronExpression != "" {
		_, err := parser.Parse(in.CronExpression)
		if err != nil {
			return &system.RespBase{
				Base: helper.GetErrResp(i18n.InvalidCronExpression, i18n.Translate(i18n.InvalidCronExpression, l.ctx)),
			}, nil
		}
	}
	job, err := l.svcCtx.JobModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.InternalServerError, i18n.Translate(i18n.InternalServerError, l.ctx)),
		}, nil
	}
	if job == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.CronJobNotFound, i18n.Translate(i18n.CronJobNotFound, l.ctx)),
		}, nil
	}
	if in.JobName != "" {
		job.JobName = in.JobName
	}
	if in.JobGroup != "" {
		job.JobGroup = in.JobGroup
	}
	if in.InvokeTarget != "" {
		job.InvokeTarget = in.InvokeTarget
	}
	if in.CronExpression != "" {
		job.CronExpression = in.CronExpression
	}
	if in.Status != 0 {
		job.Status = jobStatusToModel(in.Status)
	}
	if in.Remark != "" {
		job.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	userName, err := utils.GetUsernameFromMd(l.ctx)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.InternalServerError, i18n.Translate(i18n.InternalServerError, l.ctx)),
		}, nil
	}
	job.UpdateBy = sql.NullString{String: userName, Valid: true}
	job.UpdateTimes = utils.NowMillis()

	err = l.svcCtx.JobModel.Update(l.ctx, job)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.InternalServerError, i18n.Translate(i18n.InternalServerError, l.ctx)),
		}, nil
	}

	err = l.svcCtx.Cron.ReloadJob(job)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.InternalServerError, i18n.Translate(i18n.InternalServerError, l.ctx)),
		}, nil
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
