package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleUpdateLogic {
	return &SysRoleUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysRoleUpdateLogic) SysRoleUpdate(in *system.SysRoleUpdateReq) (*system.RespBase, error) {
	one, err := l.svcCtx.RoleModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return &system.RespBase{
			Code: 400,
			Msg:  "角色不存在",
		}, nil
	}

	var data models.SysRole
	_ = copier.Copy(&data, one)
	_ = copier.Copy(&data, in)

	err = l.svcCtx.RoleModel.Update(l.ctx, &data)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Code: 200,
		Msg:  "更新成功",
	}, nil
}
