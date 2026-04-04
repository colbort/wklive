package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysRoleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysRoleListLogic {
	return &SysRoleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 角色
func (l *SysRoleListLogic) SysRoleList(in *system.SysRoleListReq) (*system.SysRoleListResp, error) {
	// 2) 查分页
	items, total, err := l.svcCtx.RoleModel.FindPage(l.ctx, in.Keyword, in.Status, in.Page.Cursor, in.Page.Limit)
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

	// 3) 组装返回
	data := make([]*system.SysRoleItem, 0, len(items))
	for _, r := range items {
		data = append(data, &system.SysRoleItem{
			Id:          r.Id,
			Name:        r.Name,
			Code:        r.Code,
			Status:      r.Status,
			Remark:      r.Remark,
			CreateTimes: r.CreateTimes,
		})
	}

	return &system.SysRoleListResp{
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
