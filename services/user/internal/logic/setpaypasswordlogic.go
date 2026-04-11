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
)

type SetPayPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetPayPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetPayPasswordLogic {
	return &SetPayPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置支付密码
func (l *SetPayPasswordLogic) SetPayPassword(in *user.SetPayPasswordReq) (*user.AppCommonResp, error) {
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
	if in.Password != in.ConfirmPassword {
		return &user.AppCommonResp{
			Base: helper.GetErrResp(400, i18n.Translate(i18n.PasswordsDoNotMatch, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity != nil {
		// 更新现有支付密码
		userSecurity.PayPasswordHash = sql.NullString{String: in.Password, Valid: true}
		userSecurity.UpdateTimes = now

		err = l.svcCtx.UserSecurityModel.Update(l.ctx, userSecurity)
		if err != nil {
			return nil, err
		}
	} else {
		// 创建新的支付密码
		userSecurity = &models.TUserSecurity{
			Id:              l.svcCtx.Node.Generate().Int64(),
			TenantId:        tuser.TenantId,
			UserId:          in.UserId,
			PayPasswordHash: sql.NullString{String: in.Password, Valid: true},
			CreateTimes:     now,
			UpdateTimes:     now,
		}

		_, err = l.svcCtx.UserSecurityModel.Insert(l.ctx, userSecurity)
		if err != nil {
			return nil, err
		}
	}

	l.Logger.Infof("用户 %d 设置支付密码成功", in.UserId)

	return &user.AppCommonResp{
		Base: helper.OkResp(),
	}, nil
}
