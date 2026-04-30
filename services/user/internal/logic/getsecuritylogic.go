package logic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSecurityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSecurityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSecurityLogic {
	return &GetSecurityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取安全设置
func (l *GetSecurityLogic) GetSecurity(in *user.GetSecurityReq) (*user.GetSecurityResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.GetSecurityResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 查询用户安全信息
	security, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if security == nil {
		return &user.GetSecurityResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.SecuritySettingsNotFound, l.ctx)),
		}, nil
	}

	return &user.GetSecurityResp{
		Base:     helper.OkResp(),
		Security: toUserSecurityProto(security),
	}, nil
}
