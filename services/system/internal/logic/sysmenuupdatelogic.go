package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysMenuUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuUpdateLogic {
	return &SysMenuUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysMenuUpdateLogic) SysMenuUpdate(in *system.SysMenuUpdateReq) (*system.RespBase, error) {
	if base, err := systemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}

	one, err := l.svcCtx.MenuModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.MenuNotFound, i18n.Translate(i18n.MenuNotFound, l.ctx)),
		}, nil
	}

	if in.ParentId != 0 {
		one.ParentId = in.ParentId
	}
	if in.Name != "" {
		one.Name = in.Name
	}
	if in.MenuType != system.MenuType_MENU_TYPE_UNKNOWN {
		one.MenuType = menuTypeToModel(in.MenuType)
	}
	if in.Method != system.RequestMethod_REQUEST_METHOD_UNKNOWN {
		one.Method = requestMethodToString(in.Method)
	}
	if in.Path != "" {
		one.Path = in.Path
	}
	if in.Component != "" {
		one.Component = in.Component
	}
	if in.Icon != "" {
		one.Icon = in.Icon
	}
	if in.Sort != 0 {
		one.Sort = in.Sort
	}
	if in.Visible != common.Switch_SWITCH_UNKNOWN {
		one.Visible = visibleStatusToModel(in.Visible)
	}
	if in.Enabled != common.Enable_ENABLE_UNKNOWN {
		one.Enabled = commonStatusToModel(in.Enabled)
	}
	if in.Perms != "" {
		one.Perms = in.Perms
	}
	one.UpdateTimes = utils.NowMillis()

	err = l.svcCtx.MenuModel.Update(l.ctx, one)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
