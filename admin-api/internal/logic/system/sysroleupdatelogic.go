// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/rpc/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleUpdateLogic {
	return &SysRoleUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysRoleUpdateLogic) SysRoleUpdate(req *types.SysRoleUpdateReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysRoleUpdate(l.ctx, &system.SysRoleUpdateReq{
		Id:     req.Id,
		Name:   req.Name,
		Status: req.Status,
		Remark: req.Remark,
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
