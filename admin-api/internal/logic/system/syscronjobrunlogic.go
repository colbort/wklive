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

type SysCronJobRunLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobRunLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobRunLogic {
	return &SysCronJobRunLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobRunLogic) SysCronJobRun(req *types.SysCronJobRunReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysCronJobRun(l.ctx, &system.SysCronJobRunReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.RespBase{
		Code: result.Code,
		Msg:  result.Msg,
	}
	return
}
