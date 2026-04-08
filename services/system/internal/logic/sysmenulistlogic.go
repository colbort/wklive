package logic

import (
	"context"

	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysMenuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuListLogic {
	return &SysMenuListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取菜单列表
func (l *SysMenuListLogic) SysMenuList(in *system.SysMenuListReq) (*system.SysMenuListResp, error) {
	// 2) 查分页
	items, total, err := l.svcCtx.MenuModel.FindPage(l.ctx, in.Keyword, in.MenuType, in.Status, in.Visible, in.Page.Cursor, in.Page.Limit)
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
	data := make([]*system.SysMenuItem, 0, len(items))
	for _, r := range items {
		data = append(data, &system.SysMenuItem{
			Id:        r.Id,
			ParentId:  r.ParentId,
			Name:      r.Name,
			MenuType:  r.MenuType,
			Path:      r.Path,
			Component: r.Component,
			Icon:      r.Icon,
			Sort:      r.Sort,
			Visible:   r.Visible,
			Status:    r.Status,
			Perms:     r.Perms,
		})
	}
	return &system.SysMenuListResp{
		Base: &common.RespBase{
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
