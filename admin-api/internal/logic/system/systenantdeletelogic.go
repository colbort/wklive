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

type SysTenantDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysTenantDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDeleteLogic {
	return &SysTenantDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysTenantDeleteLogic) SysTenantDelete(req *types.SysTenantDeleteReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysTenantDelete(l.ctx, &system.SysTenantDeleteReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}
