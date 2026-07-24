package adminlogic

import (
	"reflect"
	"testing"

	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/models"
)

func TestBuildMenuTreeAndPermsIncludesHiddenButtons(t *testing.T) {
	rows := []*models.SysMenu{
		{
			Id:       20000,
			ParentId: 0,
			Name:     "做市管理",
			MenuType: int64(system.MenuType_MENU_TYPE_DIR),
			Visible:  int64(common.Switch_SWITCH_ON),
		},
		{
			Id:       20100,
			ParentId: 20000,
			Name:     "流动性提供方",
			MenuType: int64(system.MenuType_MENU_TYPE_MENU),
			Visible:  int64(common.Switch_SWITCH_ON),
			Perms:    "liquidity:provider:list",
		},
		{
			Id:       20101,
			ParentId: 20100,
			Name:     "新增提供方",
			MenuType: int64(system.MenuType_MENU_TYPE_BUTTON),
			Visible:  int64(common.Switch_SWITCH_OFF),
			Perms:    "liquidity:provider:add",
		},
	}

	tree, perms := buildMenuTreeAndPerms(rows)
	if len(tree) != 1 || len(tree[0].Children) != 1 {
		t.Fatalf("unexpected menu tree: %+v", tree)
	}
	if !reflect.DeepEqual(perms, []string{
		"liquidity:provider:add",
		"liquidity:provider:list",
	}) {
		t.Fatalf("unexpected perms: %v", perms)
	}
}
