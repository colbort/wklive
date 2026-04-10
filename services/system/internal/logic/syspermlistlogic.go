package logic

import (
	"context"
	"sort"
	"strings"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysPermListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysPermListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysPermListLogic {
	return &SysPermListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取权限列表
func (l *SysPermListLogic) SysPermList(in *system.Empty) (*system.SysPermListResp, error) {
	menus, err := l.svcCtx.MenuModel.ListAll(l.ctx)
	if err != nil {
		return nil, err
	}

	data := make([]*system.SysPermItem, 0)
	for _, menu := range menus {
		if int32(menu.MenuType) == 1 {
			continue
		}
		if int32(menu.Status) != 1 {
			continue
		}

		key := strings.TrimSpace(menu.Perms)
		if key == "" {
			continue
		}
		path := strings.TrimSpace(menu.Path)
		if path == "" {
			continue
		}

		method := strings.TrimSpace(menu.Method)
		if method == "" {
			continue
		}
		data = append(data, &system.SysPermItem{
			PermKey: key,
			Method:  method,
			Path:    path,
			Name:    menu.Name,
		})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].PermKey < data[j].PermKey
	})

	return &system.SysPermListResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}
