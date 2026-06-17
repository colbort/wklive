package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

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
	tenantId, _ := utils.GetTenantIdFromMd(l.ctx)
	// 2️⃣ 查询用户分页
	items, total, err := l.svcCtx.UserModel.FindPage(
		l.ctx,
		models.UserPageFilter{
			Keyword:  in.Keyword,
			TenantId: tenantId,
			Enabled:  commonStatusToModel(in.Enabled),
		},
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return &system.SysUserListResp{
			Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, 0, total, 0),
			Data: []*system.SysUserItem{},
		}, nil
	}

	// 3️⃣ 收集 userIds
	userIds := make([]int64, 0, len(items))
	for _, u := range items {
		userIds = append(userIds, u.Id)
	}

	// 4️⃣ 批量查询用户角色
	roleMap, err := l.svcCtx.UserRoleModel.FindRoleIdsByUserIds(l.ctx, userIds)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	// 5️⃣ 组装返回
	data := make([]*system.SysUserItem, 0, len(items))
	for _, u := range items {
		data = append(data, &system.SysUserItem{
			Id:               u.Id,
			Username:         u.Username,
			Nickname:         u.Nickname,
			Enabled:          commonStatusToProto(u.Enabled),
			RoleIds:          roleMap[u.Id],
			CreateTimes:      u.CreateTimes,
			Google2FaEnabled: commonStatusToProto(u.GoogleEnabled),
			TenantId:         u.TenantId,
			UserType:         system.UserType(u.UserType),
			IsOwner:          common.YesNo(u.IsOwner),
			Avatar:           u.Avatar,
			PermsVer:         u.PermsVer,
			LastLoginIp:      u.LastLoginIp,
			LastLoginAt:      u.LastLoginAt,
			CreateBy:         u.CreateBy,
			UpdateTimes:      u.UpdateTimes,
		})
	}

	return &system.SysUserListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
