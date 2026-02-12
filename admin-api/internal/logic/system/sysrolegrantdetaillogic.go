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

type SysRoleGrantDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysRoleGrantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleGrantDetailLogic {
	return &SysRoleGrantDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysRoleGrantDetailLogic) SysRoleGrantDetail(req *types.SysRoleGrantDetailReq) (resp *types.SysRoleGrantDetailResp, err error) {
	reuslt, err := l.svcCtx.SystemCli.SysRoleGrantDetail(l.ctx, &system.SysRoleGrantDetailReq{
		RoleId: req.Id,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.SysRoleGrantDetailResp{
		RespBase: types.RespBase{
			Code: reuslt.Code,
			Msg:  reuslt.Msg,
		},
		Data: types.SysRoleGrantDetail{
			RoleId:   reuslt.RoleId,
			MenuIds:  reuslt.MenuIds,
			PermKeys: reuslt.PermKeys,
		},
	}
	return
}
