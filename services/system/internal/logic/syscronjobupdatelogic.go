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
	_, err := cron.ParseStandard(in.CronExpression)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(400, i18n.Translate(i18n.InvalidCronExpression, l.ctx)),
		}, nil
	}
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
	job.JobName = in.JobName
	job.JobGroup = in.JobGroup
	job.InvokeTarget = in.InvokeTarget
	job.CronExpression = in.CronExpression
	job.Status = jobStatusToModel(in.Status)
	job.Remark = sql.NullString{String: in.Remark, Valid: true}
	userName, err := utils.GetUsernameFromCtx(l.ctx)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}
	job.UpdateBy = sql.NullString{String: userName, Valid: true}
	job.UpdateTimes = utils.NowMillis()

	err = l.svcCtx.JobModel.Update(l.ctx, job)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}

	err = l.svcCtx.Cron.ReloadJob(job)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
