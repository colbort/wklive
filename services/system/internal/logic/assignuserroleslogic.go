package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
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
		return &system.RespBase{Code: 1, Msg: "用户ID错误"}, nil
	}
	if len(in.RoleIds) == 0 {
		return &system.RespBase{Code: 1, Msg: "请选择角色"}, nil
	}

	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &system.RespBase{Code: 1, Msg: "用户不存在"}, nil
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
		return &system.RespBase{Code: 1, Msg: "请选择有效角色"}, nil
	}

	// 2. 校验这些角色是否存在
	// 这里按你的 model 自己改，比如 FindListByIds / CountByIds / FindByIds
	existRoleIds, err := l.svcCtx.RoleModel.FindIdsByIds(l.ctx, roleIds)
	if err != nil {
		return nil, err
	}
	if len(existRoleIds) != len(roleIds) {
		return &system.RespBase{Code: 1, Msg: "部分角色不存在"}, nil
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
	for _, roleId := range roleIds {
		if _, ok := currentMap[roleId]; ok {
			continue
		}
		_, err := l.svcCtx.UserRoleModel.Insert(l.ctx, &models.SysUserRole{
			UserId: in.UserId,
			RoleId: roleId,
		})
		if err != nil {
			return nil, err
		}
	}

	return &system.RespBase{}, nil
}
