package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改头像，密码，昵称
func (l *UpdateProfileLogic) UpdateProfile(in *system.UpdateProfileReq) (*system.RespBase, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
		}
		return nil, err
	}

	if user == nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}
	if in.Avatar != nil && *in.Avatar != "" {
		user.Avatar = *in.Avatar
	}
	if in.Nickname != nil && *in.Nickname != "" {
		user.Nickname = *in.Nickname
	}
	if in.Password != nil && *in.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}
	err = l.svcCtx.UserModel.Update(l.ctx, user)
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
