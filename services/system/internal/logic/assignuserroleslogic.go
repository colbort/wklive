package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AssignUserRolesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignUserRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignUserRolesLogic {
	return &AssignUserRolesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AssignUserRolesLogic) AssignUserRoles(in *system.AssignUserRolesReq) (*system.RespBase, error) {
	if in.UserId <= 0 {
		return &system.RespBase{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.UserIDInvalid, l.ctx)),
		}, nil
	}
	if len(in.RoleIds) == 0 {
		return &system.RespBase{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.RoleSelectionRequired, l.ctx)),
		}, nil
	}

	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 1. 请求角色去重
	roleMap := make(map[int64]struct{}, len(in.RoleIds))
	roleIds := make([]int64, 0, len(in.RoleIds))
	for _, roleId := range in.RoleIds {
		if roleId <= 0 {
			continue
		}
		if _, ok := roleMap[roleId]; ok {
			continue
		}
		roleMap[roleId] = struct{}{}
		roleIds = append(roleIds, roleId)
	}

	if len(roleIds) == 0 {
		return &system.RespBase{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.ValidRoleSelectionRequired, l.ctx)),
		}, nil
	}

	// 2. 校验这些角色是否存在
	// 这里按你的 model 自己改，比如 FindListByIds / CountByIds / FindByIds
	existRoleIds, err := l.svcCtx.RoleModel.FindIdsByIds(l.ctx, roleIds)
	if err != nil {
		return nil, err
	}
	if len(existRoleIds) != len(roleIds) {
		return &system.RespBase{
			Base: helper.GetErrResp(1, i18n.Translate(i18n.SomeRolesNotFound, l.ctx)),
		}, nil
	}

	// 3. 查用户当前已有角色
	currentRoleIds, err := l.svcCtx.UserRoleModel.FindRoleIdsByUserId(l.ctx, in.UserId)
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}

	currentMap := make(map[int64]struct{}, len(currentRoleIds))
	for _, roleId := range currentRoleIds {
		currentMap[roleId] = struct{}{}
	}

	// 4. 插入缺少的角色
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userRoleModel := models.NewSysUserRoleModel(conn, l.svcCtx.Config.CacheRedis).(models.UserRoleModel)

		for _, roleId := range roleIds {
			if _, ok := currentMap[roleId]; ok {
				continue
			}
			if _, err := userRoleModel.Insert(ctx, &models.SysUserRole{
				UserId: in.UserId,
				RoleId: roleId,
			}); err != nil {
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
