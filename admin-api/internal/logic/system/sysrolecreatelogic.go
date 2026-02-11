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

type SysRoleCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleCreateLogic {
	return &SysRoleCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysRoleCreateLogic) SysRoleCreate(req *types.SysRoleCreateReq) (resp *types.SimpleResp, err error) {
	result, err := l.svcCtx.SystemCli.SysRoleCreate(l.ctx, &system.SysRoleCreateReq{
		Name:   req.Name,
		Code:   req.Code,
		Status: req.Status,
		Remark: req.Remark,
	})
	return &types.SimpleResp{
		Code: int(result.Code),
		Msg:  result.Msg,
	}, err
}
