// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/rpc/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysMenuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuListLogic {
	return &SysMenuListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysMenuListLogic) SysMenuList(req *types.SysMenuListReq) (resp *types.SysMenuListResp, err error) {
	result, err := l.svcCtx.SystemCli.SysMenuList(l.ctx, &system.SysMenuListReq{
		Page: &system.PageReq{
			Page: req.Page,
			Size: req.Size,
		},
		Keyword:  req.Keyword,
		Status:   req.Status,
		Visible:  req.Visible,
		MenuType: req.MenuType,
	})
	if err != nil {
		return nil, err
	}
	data := make([]types.SysMenuItem, 0)
	for _, v := range result.Data {
		data = append(data, types.SysMenuItem{
			Id:        v.Id,
			ParentId:  v.ParentId,
			Name:      v.Name,
			MenuType:  v.MenuType,
			Path:      v.Path,
			Component: v.Component,
			Icon:      v.Icon,
			Sort:      v.Sort,
			Visible:   v.Visible,
			Status:    v.Status,
			Perms:     v.Perms,
		})
	}
	return &types.SysMenuListResp{
		RespBase: types.RespBase{
			Code:  result.Base.Code,
			Msg:   result.Base.Msg,
			Total: result.Base.Total,
		},
		Data: data,
	}, nil
}
