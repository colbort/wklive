package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"
)

type UpdateUserLevelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLevelLogic {
	return &UpdateUserLevelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户会员等级
func (l *UpdateUserLevelLogic) UpdateUserLevel(in *user.UpdateUserLevelReq) (*user.AdminCommonResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.AdminCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}
	if tuser.TenantId != in.TenantId {
		return &user.AdminCommonResp{
			Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionOperateThisUser, l.ctx)),
		}, nil
	}

	// 更新会员等级
	if in.MemberLevel != 0 {
		tuser.MemberLevel = in.MemberLevel
	}
	tuser.UpdateTimes = time.Now().UnixMilli()

	err = l.svcCtx.UserModel.Update(l.ctx, tuser)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新用户 %d 会员等级为 %d", in.UserId, in.MemberLevel)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
