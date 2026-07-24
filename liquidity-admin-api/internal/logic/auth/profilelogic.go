// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth

import (
	"context"
	"strings"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileLogic {
	return &ProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProfileLogic) Profile() (resp *types.ProfileResp, err error) {
	out, err := l.svcCtx.SystemCli.GetProfile(l.ctx, &system.Empty{})
	if err != nil {
		return nil, err
	}
	resp = logicutil.Convert[types.ProfileResp](out)
	resp.Data.User.AppScope = int32(system.ApplicationScope_APPLICATION_SCOPE_LIQUIDITY)
	resp.Data.Menus = liquidityMenus(resp.Data.Menus)
	resp.Data.Perms = liquidityPerms(resp.Data.Perms)
	return resp, nil
}

func liquidityMenus(menus []types.MenuNode) []types.MenuNode {
	result := make([]types.MenuNode, 0, len(menus))
	for _, menu := range menus {
		menu.Children = liquidityMenus(menu.Children)
		path := strings.TrimSpace(menu.Path)
		if path == "/liquidity" ||
			strings.HasPrefix(path, "/liquidity/") ||
			strings.HasPrefix(path, "/admin/liquidity/") ||
			len(menu.Children) > 0 {
			result = append(result, menu)
		}
	}
	return result
}

func liquidityPerms(perms []string) []string {
	result := make([]string, 0, len(perms))
	for _, perm := range perms {
		if strings.HasPrefix(perm, "liquidity:") {
			result = append(result, perm)
		}
	}
	return result
}
