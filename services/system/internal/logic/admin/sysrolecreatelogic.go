package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
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
	scope := normalizeApplicationScope(in.AppScope)
	result, err := l.svcCtx.RoleModel.FindOneByTenantIdAppScopeCode(l.ctx, in.TenantId, int64(scope), in.Code)
	if result != nil {
		return nil, i18n.StatusError(l.ctx, i18n.RoleCodeAlreadyExists)
	}
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}
	_, err = l.svcCtx.RoleModel.Insert(l.ctx, &models.SysRole{
		TenantId:    in.TenantId,
		AppScope:    int64(scope),
		Name:        in.Name,
		Code:        in.Code,
		Enabled:     commonStatusToModel(in.Enabled),
		Remark:      in.Remark,
		CreateTimes: utils.NowMillis(),
		UpdateTimes: utils.NowMillis(),
	})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
