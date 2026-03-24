package logic

import (
	"context"
	"database/sql"
	"time"

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
			Code: 400,
			Msg:  "无效的 Cron 表达式",
		}, nil
	}
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
	job.JobName = in.JobName
	job.JobGroup = in.JobGroup
	job.InvokeTarget = in.InvokeTarget
	job.CronExpression = in.CronExpression
	job.Status = in.Status
	job.Remark = sql.NullString{String: in.Remark, Valid: true}
	userName, err := utils.GetUsernameFromCtx(l.ctx)
	if err != nil {
		return &system.RespBase{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}
	job.UpdateBy = sql.NullString{String: userName, Valid: true}
	job.UpdateTime = time.Now()

	err = l.svcCtx.JobModel.Update(l.ctx, job)
	if err != nil {
		return &system.RespBase{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	err = l.svcCtx.Cron.ReloadJob(job)
	if err != nil {
		return &system.RespBase{
			Code: 500,
			Msg:  err.Error(),
		}, nil
	}

	return &system.RespBase{
		Code: 200,
		Msg:  "更新成功",
	}, nil
}
