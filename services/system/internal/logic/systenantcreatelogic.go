package logic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"strings"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type SysTenantCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantCreateLogic {
	return &SysTenantCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建租户
func (l *SysTenantCreateLogic) SysTenantCreate(in *system.SysTenantCreateReq) (*system.RespBase, error) {
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err == nil && tenantId > 0 {
		return &system.RespBase{
			Base: i18n.BuildPermissionDeniedErrorResponse(l.ctx, ""),
		}, nil
	}
	creatorId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		creatorId = 0
	}
	creatorName, err := utils.GetUsernameFromMd(l.ctx)
	if err != nil {
		creatorName = ""
	}

	tenantCode, err := l.generateTenantCode()
	if err != nil {
		return nil, err
	}

	const ownerRoleTemplateId int64 = 2
	const ownerRoleCode = "tenant_owner"
	const ownerRoleName = "租户主账号"

	now := utils.NowMillis()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.TenantPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		tenantModel := models.NewSysTenantModel(conn, l.svcCtx.Config.CacheRedis).(models.TenantModel)
		roleModel := models.NewSysRoleModel(conn, l.svcCtx.Config.CacheRedis).(models.RoleModel)
		userModel := models.NewSysUserModel(conn, l.svcCtx.Config.CacheRedis).(models.UserModel)
		userRoleModel := models.NewSysUserRoleModel(conn, l.svcCtx.Config.CacheRedis).(models.UserRoleModel)
		roleMenuModel := models.NewSysRoleMenuModel(conn, l.svcCtx.Config.CacheRedis).(models.RoleMenuModel)

		// 1. 创建租户
		tenantRes, err := tenantModel.Insert(ctx, &models.SysTenant{
			TenantCode:   tenantCode,
			TenantName:   in.TenantName,
			Enabled:      commonStatusToModel(in.Enabled),
			ExpireTime:   in.ExpireTime,
			ContactName:  sql.NullString{String: in.ContactName, Valid: in.ContactName != ""},
			ContactPhone: sql.NullString{String: in.ContactPhone, Valid: in.ContactPhone != ""},
			LoginIp:      sql.NullString{},
			LoginTime:    utils.NowMillis(),
			LoginCount:   0,
			Remark:       sql.NullString{String: in.Remark, Valid: in.Remark != ""},
			CreateBy:     sql.NullString{String: creatorName, Valid: creatorName != ""},
			CreateTimes:  now,
			UpdateBy:     sql.NullString{String: creatorName, Valid: creatorName != ""},
			UpdateTimes:  now,
		})
		if err != nil {
			return err
		}
		newTenantId, err := tenantRes.LastInsertId()
		if err != nil {
			return err
		}

		// 2. 创建租户默认最高权限角色
		roleRes, err := roleModel.Insert(ctx, &models.SysRole{
			TenantId:    newTenantId,
			Name:        ownerRoleName,
			Code:        ownerRoleCode,
			Enabled:     commonStatusToModel(in.Enabled),
			Remark:      "system bootstrap tenant owner role",
			CreateTimes: now,
			UpdateTimes: now,
		})
		if err != nil {
			return err
		}
		newRoleId, err := roleRes.LastInsertId()
		if err != nil {
			return err
		}

		// 3. 复制系统模板角色菜单到租户角色
		templateRoleMenus, err := l.svcCtx.RoleMenuModel.ListByRoleId(ctx, ownerRoleTemplateId)
		if err != nil {
			return err
		}
		if len(templateRoleMenus) == 0 {
			return i18n.StatusError(ctx, i18n.RoleNotFound)
		}
		for _, item := range templateRoleMenus {
			if _, err = roleMenuModel.Insert(ctx, &models.SysRoleMenu{
				TenantId: newTenantId,
				RoleId:   newRoleId,
				MenuId:   item.MenuId,
			}); err != nil {
				return err
			}
		}

		// 4. 创建租户主账号
		userRes, err := userModel.InsertCtx(ctx, session, &models.SysUser{
			TenantId:      newTenantId,
			UserType:      int64(system.UserType_USER_TYPE_TENANT_OWNER),
			IsOwner:       int64(common.YesNo_YES_NO_YES),
			Username:      in.Username,
			Password:      string(hashedPassword),
			Nickname:      in.TenantName,
			Avatar:        "",
			Enabled:       commonStatusToModel(in.Enabled),
			GoogleSecret:  "",
			GoogleEnabled: int64(common.Enable_ENABLE_DISABLED),
			PermsVer:      1,
			LastLoginIp:   "",
			LastLoginAt:   0,
			CreateBy:      creatorId,
			CreateTimes:   now,
			UpdateTimes:   now,
		})
		if err != nil {
			return err
		}
		newUserId, err := userRes.LastInsertId()
		if err != nil {
			return err
		}

		// 5. 绑定租户主账号角色
		_, err = userRoleModel.InsertCtx(ctx, session, &models.SysUserRole{
			TenantId: newTenantId,
			UserId:   newUserId,
			RoleId:   newRoleId,
		})
		return err
	})
	if err != nil {
		if err.Error() == "tenant_owner_role_template_not_found" {
			return &system.RespBase{
				Base: helper.GetErrResp(i18n.TenantOwnerRoleTemplateNotFound, i18n.Translate(i18n.TenantOwnerRoleTemplateNotFound, l.ctx)),
			}, nil
		}
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

func (l *SysTenantCreateLogic) generateTenantCode() (string, error) {
	const maxAttempts = 20

	for i := 0; i < maxAttempts; i++ {
		code, err := randomAlphaNum(5)
		if err != nil {
			return "", err
		}

		tenant, err := l.svcCtx.TenantMode.FindByTenantCode(l.ctx, code)
		if errors.Is(err, models.ErrNotFound) {
			return code, nil
		}
		if err != nil {
			return "", err
		}
		if tenant == nil {
			return code, nil
		}
	}

	return "", i18n.StatusError(l.ctx, i18n.InternalServerError)
}

func randomAlphaNum(length int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	buf := make([]byte, length)
	max := big.NewInt(int64(len(letters)))
	for i := range buf {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		buf[i] = letters[n.Int64()]
	}

	return strings.ToUpper(string(buf)), nil
}
