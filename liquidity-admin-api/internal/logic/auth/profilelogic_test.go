package auth

import (
	"reflect"
	"testing"

	"wklive/liquidity-admin-api/internal/types"
)

func TestLiquidityProfileFiltering(t *testing.T) {
	menus := []types.MenuNode{
		{Id: 1, Path: "/system/users"},
		{Id: 2, Path: "/liquidity", Children: []types.MenuNode{
			{Id: 3, Path: "/liquidity/admin/providers"},
		}},
	}
	gotMenus := liquidityMenus(menus)
	if len(gotMenus) != 1 || gotMenus[0].Id != 2 || len(gotMenus[0].Children) != 1 {
		t.Fatalf("unexpected menus: %+v", gotMenus)
	}
	gotPerms := liquidityPerms([]string{"sys:user:list", "liquidity:provider:list"})
	if !reflect.DeepEqual(gotPerms, []string{"liquidity:provider:list"}) {
		t.Fatalf("unexpected perms: %v", gotPerms)
	}
}
