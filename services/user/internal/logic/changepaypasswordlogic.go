package logic

import (
	"context"
	"database/sql"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"
)

type ChangePayPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePayPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePayPasswordLogic {
	return &ChangePayPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改支付密码
func (l *ChangePayPasswordLogic) ChangePayPassword(in *user.ChangePayPasswordReq) (*user.AppCommonResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 验证密码是否一致
	if in.NewPassword != in.ConfirmPassword {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(400, i18n.Translate(i18n.PasswordsDoNotMatch, l.ctx)),
		}, nil
	}

	// 获取用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity == nil {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.PayPasswordNotSet, l.ctx)),
		}, nil
	}

	// TODO: 验证旧密码是否正确
	// 在实际项目中需要对密码进行验证

	// 更新支付密码
	userSecurity.PayPasswordHash = sql.NullString{String: in.NewPassword, Valid: true}
	userSecurity.UpdateTimes = time.Now().UnixMilli()

	err = l.svcCtx.UserSecurityModel.Update(l.ctx, userSecurity)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 修改支付密码成功", in.UserId)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
