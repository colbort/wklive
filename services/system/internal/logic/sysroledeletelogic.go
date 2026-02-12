package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleDeleteLogic {
	return &SysRoleDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysRoleDeleteLogic) SysRoleDelete(in *system.SysRoleDeleteReq) (*system.RespBase, error) {
	err := l.svcCtx.RoleModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Code: 200,
		Msg:  "删除成功",
	}, nil
}
