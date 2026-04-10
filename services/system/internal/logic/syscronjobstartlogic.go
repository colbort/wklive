package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
			Base: &common.RespBase{
				Code: 500,
				Msg:  err.Error(),
			},
		}, nil
	}
	if job == nil {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 400,
				Msg:  "定时任务不存在",
			},
		}, nil
	}
	err = l.svcCtx.Cron.StartJob(job)
	if err != nil {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 500,
				Msg:  err.Error(),
			},
		}, nil
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
