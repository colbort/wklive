package adminlogic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysMenuCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysMenuCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysMenuCreateLogic {
	return &SysMenuCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 菜单
func (l *SysMenuCreateLogic) SysMenuCreate(in *system.SysMenuCreateReq) (*system.RespBase, error) {
	scope := normalizeApplicationScope(in.AppScope)
	if in.ParentId > 0 {
		parent, findErr := l.svcCtx.MenuModel.FindOne(l.ctx, in.ParentId)
		if findErr != nil {
			return nil, findErr
		}
		if parent.AppScope != int64(scope) {
			return nil, i18n.StatusError(l.ctx, i18n.Forbidden)
		}
	}
	var menu *models.SysMenu
	var err error
	switch menuTypeToModel(in.MenuType) {
	case 1:
		menu, err = l.svcCtx.MenuModel.FindOneByName(l.ctx, int64(scope), in.Name)
	case 2:
		menu, err = l.svcCtx.MenuModel.FindOneByPath(l.ctx, int64(scope), in.Path)
	case 3:
		menu, err = l.svcCtx.MenuModel.FindOneByPerms(l.ctx, int64(scope), in.Perms)
	default:
		return &system.RespBase{
			Base: helper.ErrResp(i18n.InvalidMenuType, i18n.Translate(i18n.InvalidMenuType, l.ctx)),
		}, nil
	}
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}
	if menu != nil {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.MenuAlreadyExists, i18n.Translate(i18n.MenuAlreadyExists, l.ctx)),
		}, nil
	}
	_, err = l.svcCtx.MenuModel.Insert(l.ctx, &models.SysMenu{
		ParentId:    in.ParentId,
		AppScope:    int64(scope),
		Name:        in.Name,
		MenuType:    menuTypeToModel(in.MenuType),
		Method:      requestMethodToString(in.Method),
		Path:        in.Path,
		Component:   in.Component,
		Perms:       in.Perms,
		Icon:        in.Icon,
		Sort:        in.Sort,
		Visible:     visibleStatusToModel(in.Visible),
		Enabled:     commonStatusToModel(in.Enabled),
		CreateTimes: utils.NowMillis(),
		UpdateTimes: utils.NowMillis(),
	})
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
