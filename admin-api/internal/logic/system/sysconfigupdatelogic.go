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

type SysConfigUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigUpdateLogic {
	return &SysConfigUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysConfigUpdateLogic) SysConfigUpdate(req *types.SysConfigUpdateReq) (resp *types.RespBase, err error) {
	out, err := l.svcCtx.SystemCli.SysConfigUpdate(l.ctx, &system.SysConfigUpdateReq{
		Id:          req.Id,
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.RespBase{
		Code: out.Code,
		Msg:  out.Msg,
	}
	return
}
