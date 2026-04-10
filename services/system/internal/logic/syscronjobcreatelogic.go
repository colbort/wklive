package logic

import (
	"context"
	"database/sql"
	"time"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysCronJobCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobCreateLogic {
	return &SysCronJobCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建系统定时任务
func (l *SysCronJobCreateLogic) SysCronJobCreate(in *system.SysCronJobCreateReq) (*system.RespBase, error) {
	_, err := cron.ParseStandard(in.CronExpression)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(400, "无效的 Cron 表达式"),
		}, nil
	}
	job, err := l.svcCtx.JobModel.FindByInvokeTarget(l.ctx, in.InvokeTarget)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}
	if job != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(400, "定时任务已存在"),
		}, nil
	}
	userName, err := utils.GetUsernameFromCtx(l.ctx)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}
	_, err = l.svcCtx.JobModel.Insert(l.ctx, &models.SysJob{
		JobName:        in.JobName,
		JobGroup:       in.JobGroup,
		InvokeTarget:   in.InvokeTarget,
		CronExpression: in.CronExpression,
		Status:         in.Status,
		Remark:         sql.NullString{String: in.Remark, Valid: true},
		CreateBy:       sql.NullString{String: userName, Valid: true},
		CreateTimes:    time.Now().UnixMilli(),
	})
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
