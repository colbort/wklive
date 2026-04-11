package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
)

type SysCronJobStartLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysCronJobStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobStartLogic {
	return &SysCronJobStartLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 开始系统定时任务
func (l *SysCronJobStartLogic) SysCronJobStart(in *system.SysCronJobStartReq) (*system.RespBase, error) {
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
	err = l.svcCtx.Cron.StartJob(job)
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
