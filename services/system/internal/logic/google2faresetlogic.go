package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type Google2FAResetLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGoogle2FAResetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Google2FAResetLogic {
	return &Google2FAResetLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *Google2FAResetLogic) Google2FAReset(in *system.Google2FAResetReq) (*system.RespBase, error) {
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, user.TenantId, i18n.NoPermissionOperateThisUser); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}
	user.GoogleSecret = ""
	user.GoogleEnabled = int64(common.Enable_ENABLE_DISABLED)
	if err = l.svcCtx.UserModel.Update(l.ctx, user); err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
