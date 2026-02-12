package logic

import (
	"context"
	"sort"
	"strings"

	"wklive/rpc/system"
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
	m := make(map[string]string, 256)

	for _, menu := range menus {
		if int32(menu.MenuType) != 3 {
			continue
		}
		if int32(menu.Status) != 1 {
			continue
		}

		key := strings.TrimSpace(menu.Perms)
		if key == "" {
			continue
		}
		name := strings.TrimSpace(menu.Name)
		if name == "" {
			name = key
		}

		if old, ok := m[key]; !ok || old == "" {
			m[key] = name
		}
	}

	data := make([]*system.SysPermItem, 0, len(m))
	for k, v := range m {
		data = append(data, &system.SysPermItem{
			PermKey: k,
			Name:    v,
		})
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].PermKey < data[j].PermKey
	})

	return &system.SysPermListResp{
		Base: &system.RespBase{
			Code: 200,
			Msg:  "success",
		},
		Data: data,
	}, nil
}
