package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/jinzhu/copier"
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
	one, err := l.svcCtx.MenuModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.MenuNotFound, i18n.Translate(i18n.MenuNotFound, l.ctx)),
		}, nil
	}

	var data models.SysMenu
	_ = copier.Copy(&data, one)
	copyNonZero(&data, in)
	if in.MenuType != system.MenuType_MENU_TYPE_UNKNOWN {
		data.MenuType = menuTypeToModel(in.MenuType)
	}
	if in.Method != system.RequestMethod_REQUEST_METHOD_UNKNOWN {
		data.Method = requestMethodToString(in.Method)
	}
	if in.Visible != 0 {
		data.Visible = visibleStatusToModel(in.Visible)
	}
	if in.Enabled != 0 {
		data.Enabled = commonStatusToModel(in.Enabled)
	}

	err = l.svcCtx.MenuModel.Update(l.ctx, &data)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
