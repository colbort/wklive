package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"
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
		return nil, errors.New(i18n.Translate(i18n.RoleCodeAlreadyExists, l.ctx))
	}
	_, err = l.svcCtx.RoleModel.Insert(l.ctx, &models.SysRole{
		Name:   in.Name,
		Code:   in.Code,
		Status: commonStatusToModel(in.Status),
		Remark: in.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
