package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SysTenantDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDeleteLogic {
	return &SysTenantDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除租户
func (l *SysTenantDeleteLogic) SysTenantDelete(in *system.SysTenantDeleteReq) (*system.RespBase, error) {
	tenant, err := l.svcCtx.TenantMode.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || tenant == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.TenantNotFound, i18n.Translate(i18n.TenantNotFound, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		tenantModel := models.NewSysTenantModel(conn, l.svcCtx.Config.CacheRedis).(models.TenantModel)
		userModel := models.NewSysUserModel(conn, l.svcCtx.Config.CacheRedis).(models.UserModel)
		roleModel := models.NewSysRoleModel(conn, l.svcCtx.Config.CacheRedis).(models.RoleModel)
		userRoleModel := models.NewSysUserRoleModel(conn, l.svcCtx.Config.CacheRedis).(models.UserRoleModel)
		roleMenuModel := models.NewSysRoleMenuModel(conn, l.svcCtx.Config.CacheRedis).(models.RoleMenuModel)

		roleMenuIds, err := roleMenuModel.FindIdsByTenantId(ctx, tenant.Id)
		if err != nil {
			return err
		}
		for _, id := range roleMenuIds {
			if err = roleMenuModel.Delete(ctx, id); err != nil && !errors.Is(err, models.ErrNotFound) {
				return err
			}
		}

		userRoleIds, err := userRoleModel.FindIdsByTenantId(ctx, tenant.Id)
		if err != nil {
			return err
		}
		for _, id := range userRoleIds {
			if err = userRoleModel.Delete(ctx, id); err != nil && !errors.Is(err, models.ErrNotFound) {
				return err
			}
		}

		roleIds, err := roleModel.FindIdsByTenantId(ctx, tenant.Id)
		if err != nil {
			return err
		}
		for _, id := range roleIds {
			if err = roleModel.Delete(ctx, id); err != nil && !errors.Is(err, models.ErrNotFound) {
				return err
			}
		}

		userIds, err := userModel.FindIdsByTenantId(ctx, tenant.Id)
		if err != nil {
			return err
		}
		for _, id := range userIds {
			if err = userModel.Delete(ctx, id); err != nil && !errors.Is(err, models.ErrNotFound) {
				return err
			}
		}

		return tenantModel.Delete(ctx, tenant.Id)
	})
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
