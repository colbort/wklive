package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
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
	result, err := l.svcCtx.RoleModel.FindOneByTenantIdCode(l.ctx, in.TenantId, in.Code)
	if err == nil {
		return nil, err
	}
	if result != nil {
		return nil, i18n.StatusError(l.ctx, i18n.RoleCodeAlreadyExists)
	}
	_, err = l.svcCtx.RoleModel.Insert(l.ctx, &models.SysRole{
		TenantId: in.TenantId,
		Name:     in.Name,
		Code:     in.Code,
		Enabled:  commonStatusToModel(in.Enabled),
		Remark:   in.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
