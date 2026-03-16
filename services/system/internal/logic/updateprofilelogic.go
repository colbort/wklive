package logic

import (
	"context"
	"errors"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}
	if in.Avatar != nil {
		user.Avatar = *in.Avatar
	}
	if in.Nickname != nil {
		user.Nickname = *in.Nickname
	}
	if in.Password != nil {
		user.Password = *in.Password
	}
	err = l.svcCtx.UserModel.Update(l.ctx, user)
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Code: 200,
		Msg:  "success",
	}, nil
}
