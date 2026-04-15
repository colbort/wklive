package logic

import (
	"context"

	"wklive/common/pageutil"
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
	items, total, err := l.svcCtx.MenuModel.FindPage(l.ctx, in.Keyword, menuTypeToModel(in.MenuType), commonStatusToModel(in.Status), visibleStatusToModel(in.Visible), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	// 3) 组装返回
	data := make([]*system.SysMenuItem, 0, len(items))
	for _, r := range items {
		data = append(data, &system.SysMenuItem{
			Id:        r.Id,
			ParentId:  r.ParentId,
			Name:      r.Name,
			MenuType:  menuTypeToProto(r.MenuType),
			Path:      r.Path,
			Component: r.Component,
			Icon:      r.Icon,
			Sort:      r.Sort,
			Visible:   visibleStatusToProto(r.Visible),
			Status:    commonStatusToProto(r.Status),
			Perms:     r.Perms,
		})
	}
	return &system.SysMenuListResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}
