package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type SysUserCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserCreateLogic {
	return &SysUserCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysUserCreateLogic) SysUserCreate(in *system.SysUserCreateReq) (*system.RespBase, error) {
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		tenantId = 0
	}
	creatorId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		creatorId = 0
	}

	one, err := l.svcCtx.UserModel.FindOneByTenantIdUsername(l.ctx, tenantId, in.Username)
	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}
	if one != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.UsernameAlreadyExists, i18n.Translate(i18n.UsernameAlreadyExists, l.ctx)),
		}, nil
	}

	if in.Password == "" {
		return nil, i18n.StatusError(l.ctx, i18n.ParamError)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userType := int64(system.UserType_USER_TYPE_SYSTEM_ADMIN)
	if tenantId > 0 {
		userType = int64(system.UserType_USER_TYPE_TENANT_ADMIN)
	}
	data := models.SysUser{
		TenantId:      tenantId,
		UserType:      userType,
		IsOwner:       int64(common.YesNo_YES_NO_NO),
		Username:      in.Username,
		Nickname:      in.Nickname,
		Password:      string(hashedPassword),
		PermsVer:      1,
		Enabled:       commonStatusToModel(in.Enabled),
		Avatar:        in.Avatar,
		GoogleSecret:  "",
		GoogleEnabled: int64(common.Enable_ENABLE_DISABLED),
		LastLoginIp:   "",
		LastLoginAt:   0,
		CreateBy:      creatorId,
		CreateTimes:   utils.NowMillis(),
		UpdateTimes:   utils.NowMillis(),
	}

	roleIds := make([]int64, 0, len(in.RoleIds))
	seen := make(map[int64]struct{}, len(in.RoleIds))
	for _, id := range in.RoleIds {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		roleIds = append(roleIds, id)
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userModel := models.NewSysUserModel(conn, l.svcCtx.Config.CacheRedis)
		userRoleModel := models.NewSysUserRoleModel(conn, l.svcCtx.Config.CacheRedis)

		res, err := userModel.InsertCtx(ctx, session, &data)
		if err != nil {
			return err
		}
		userId, err := res.LastInsertId()
		if err != nil {
			return err
		}
		data.Id = userId

		if len(roleIds) == 0 {
			return nil
		}

		for _, rid := range roleIds {
			role, err := l.svcCtx.RoleModel.FindOne(ctx, rid)
			if err != nil && err != sqlc.ErrNotFound {
				return err
			}
			if role == nil {
				return i18n.StatusError(ctx, i18n.RoleNotFound)
			}
			if role.TenantId != tenantId {
				return i18n.StatusError(ctx, i18n.RoleNotFound)
			}

			_, err = userRoleModel.InsertCtx(ctx, session, &models.SysUserRole{
				TenantId: tenantId,
				UserId:   data.Id,
				RoleId:   rid,
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
