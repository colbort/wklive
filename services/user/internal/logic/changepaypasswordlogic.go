package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
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
		return &user.AppCommonResp{
			Base: helper.GetErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	// 验证密码是否一致
	if in.NewPassword != in.ConfirmPassword {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(i18n.PasswordsDoNotMatch, i18n.Translate(i18n.PasswordsDoNotMatch, l.ctx)),
		}, nil
	}

	// 获取用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity == nil {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(i18n.PayPasswordNotSet, i18n.Translate(i18n.PayPasswordNotSet, l.ctx)),
		}, nil
	}

	if in.NewPassword == "" {
		return nil, i18n.StatusError(l.ctx, i18n.ParamError)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// TODO: 验证旧密码是否正确
	// 在实际项目中需要对密码进行验证

	// 更新支付密码
	userSecurity.PayPasswordHash = sql.NullString{String: string(hashedPassword), Valid: true}
	userSecurity.UpdateTimes = utils.NowMillis()

	err = l.svcCtx.UserSecurityModel.Update(l.ctx, userSecurity)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("用户 %d 修改支付密码成功", userId)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
