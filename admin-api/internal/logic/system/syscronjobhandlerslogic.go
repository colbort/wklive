// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCronJobHandlersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobHandlersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobHandlersLogic {
	return &SysCronJobHandlersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobHandlersLogic) SysCronJobHandlers() (resp *types.SysCronJobHandlersResp, err error) {
	result, err := l.svcCtx.SystemCli.SysCronJobHandlers(l.ctx, &system.Empty{})
	if err != nil {
		return nil, err
	}
	data := make([]types.SysCronJobHandler, 0)
	for _, item := range result.Data {
		data = append(data, types.SysCronJobHandler{
			InvokeTarget: item.InvokeTarget,
			JobName:      item.JobName,
		})
	}
	resp = &types.SysCronJobHandlersResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}
	return
}
