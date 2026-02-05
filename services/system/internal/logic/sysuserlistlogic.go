package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserListLogic {
	return &SysUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户
func (l *SysUserListLogic) SysUserList(in *system.SysUserListReq) (*system.SysUserListResp, error) {
	// 1️⃣ 参数兜底
	page := in.Page.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.Page.Size
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	// 2️⃣ 查询用户分页
	users, total, err := l.svcCtx.UserModel.FindPage(
		l.ctx,
		in.Keyword,
		in.Status,
		page,
		pageSize,
	)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return &system.SysUserListResp{
			Total: total,
			List:  []*system.SysUserItem{},
		}, nil
	}

	// 3️⃣ 收集 userIds
	userIds := make([]int64, 0, len(users))
	for _, u := range users {
		userIds = append(userIds, u.Id)
	}

	// 4️⃣ 批量查询用户角色
	roleMap, err := l.svcCtx.UserRoleModel.FindRoleIdsByUserIds(l.ctx, userIds)
	if err != nil {
		return nil, err
	}
	// roleMap: map[userId][]roleId

	// 5️⃣ 组装返回
	list := make([]*system.SysUserItem, 0, len(users))
	for _, u := range users {
		list = append(list, &system.SysUserItem{
			Id:               u.Id,
			Username:         u.Username,
			Nickname:         u.Nickname,
			Status:           u.Status,
			RoleIds:          roleMap[u.Id],
			CreatedAt:        u.CreatedAt.UnixMilli(),
			Google2FaEnabled: u.GoogleEnabled,
		})
	}

	return &system.SysUserListResp{
		Total: total,
		List:  list,
	}, nil
}
