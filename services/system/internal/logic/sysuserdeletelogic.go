package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SysUserDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserDeleteLogic {
	return &SysUserDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysUserDeleteLogic) SysUserDelete(in *system.SysUserDeleteReq) (*system.RespBase, error) {
	if in.Id == 1 {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.SuperAdminCannotBeDeleted, i18n.Translate(i18n.SuperAdminCannotBeDeleted, l.ctx)),
		}, nil
	}
	one, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	if one.IsOwner == int64(common.YesNo_YES_NO_YES) {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.TenantOwnerCannotBeDeleted, i18n.Translate(i18n.TenantOwnerCannotBeDeleted, l.ctx)),
		}, nil
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, one.TenantId, i18n.NoPermissionOperateThisUser); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userModel := models.NewSysUserModel(conn, l.svcCtx.Config.CacheRedis)
		userRoleModel := models.NewSysUserRoleModel(conn, l.svcCtx.Config.CacheRedis)

		roleIds, err := userRoleModel.FindRoleIdsByUserId(ctx, in.Id)
		if err != nil {
			return err
		}
		for _, roleId := range roleIds {
			userRole, err := userRoleModel.FindOneByUserIdRoleId(ctx, in.Id, roleId)
			if errors.Is(err, models.ErrNotFound) {
				continue
			}
			if err != nil {
				return err
			}
			if err = userRoleModel.Delete(ctx, userRole.Id); err != nil {
				return err
			}
		}

		return userModel.Delete(ctx, in.Id)
	})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
