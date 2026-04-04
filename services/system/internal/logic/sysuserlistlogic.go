package logic

import (
	"context"

	"wklive/proto/system"
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
	// 2️⃣ 查询用户分页
	items, total, err := l.svcCtx.UserModel.FindPage(
		l.ctx,
		in.Keyword,
		in.Status,
		in.Page.Cursor,
		in.Page.Limit,
	)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return &system.SysUserListResp{
			Base: &system.RespBase{
				Code:  200,
				Msg:   "success",
				Total: total,
			},
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

	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit

	// 5️⃣ 组装返回
	data := make([]*system.SysUserItem, 0, len(items))
	for _, u := range items {
		data = append(data, &system.SysUserItem{
			Id:               u.Id,
			Username:         u.Username,
			Nickname:         u.Nickname,
			Status:           u.Status,
			RoleIds:          roleMap[u.Id],
			CreateTimes:      u.CreateTimes,
			Google2FaEnabled: u.GoogleEnabled,
		})
	}

	return &system.SysUserListResp{
		Base: &system.RespBase{
			Code:       200,
			Msg:        "success",
			Total:      total,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
			NextCursor: nextCursor,
			PrevCursor: prevCursor,
		},
		Data: data,
	}, nil
}
