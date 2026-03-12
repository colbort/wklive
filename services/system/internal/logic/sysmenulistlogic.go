package logic

import (
	"context"

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
	// 1) 分页兜底
	page := in.Page.Page
	if page <= 0 {
		page = 1
	}
	pageSize := in.Page.Size
	if pageSize <= 0 || pageSize > 10000 {
		pageSize = 10000
	}

	// 2) 查分页
	rows, total, err := l.svcCtx.MenuModel.FindPage(l.ctx, in.Keyword, in.MenuType, in.Status, in.Visible, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 3) 组装返回
	data := make([]*system.SysMenuItem, 0, len(rows))
	for _, r := range rows {
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
		Base: &system.RespBase{
			Code:  200,
			Msg:   "success",
			Total: total,
		},
		Data: data,
	}, nil
}
