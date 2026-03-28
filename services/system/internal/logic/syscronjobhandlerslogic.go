package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobHandlersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysCronJobHandlersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobHandlersLogic {
	return &SysCronJobHandlersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取系统支持的定时任务处理器列表
func (l *SysCronJobHandlersLogic) SysCronJobHandlers(in *system.Empty) (*system.SysCronJobHandlersResp, error) {
	data := make([]*system.SysCronJobHander, 0)
	handlers := cronx.GetRegisteredNames()
	for invokeTarget, jobName := range handlers {
		data = append(data, &system.SysCronJobHander{
			InvokeTarget: invokeTarget,
			JobName:      jobName,
		})
	}
	return &system.SysCronJobHandlersResp{
		Base: &system.RespBase{
			Code: 200,
			Msg:  "success",
		},
		Data: data,
	}, nil
}
