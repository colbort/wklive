package logic

import (
	"context"
	"errors"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleCreateLogic {
	return &SysRoleCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysRoleCreateLogic) SysRoleCreate(in *system.SysRoleCreateReq) (*system.RespBase, error) {
	result, err := l.svcCtx.RoleModel.FindOneByCode(l.ctx, in.Code)
	if err == nil {
		return nil, err
	}
	if result != nil {
		return nil, errors.New("角色编码已存在")
	}
	_, err = l.svcCtx.RoleModel.Insert(l.ctx, &models.SysRole{
		Name:   in.Name,
		Code:   in.Code,
		Status: in.Status,
		Remark: in.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Code: 200,
		Msg:  "创建成功",
	}, nil
}
