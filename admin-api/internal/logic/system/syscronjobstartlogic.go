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

type SysCronJobStartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysCronJobStartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCronJobStartLogic {
	return &SysCronJobStartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCronJobStartLogic) SysCronJobStart(req *types.SysCronJobStartReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysCronJobStart(l.ctx, &system.SysCronJobStartReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}
	return
}
