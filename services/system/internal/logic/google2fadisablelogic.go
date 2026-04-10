package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type Google2FADisableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGoogle2FADisableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Google2FADisableLogic {
	return &Google2FADisableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *Google2FADisableLogic) Google2FADisable(in *system.Google2FADisableReq) (*system.RespBase, error) {
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 1,
				Msg:  "用户不存在",
			},
		}, nil
	}
	if in.Code != "" {
		if user.GoogleSecret == "" || !utils.VerifyGoogle2FACode(user.GoogleSecret, in.Code) {
			return &system.RespBase{
				Base: &common.RespBase{
					Code: 1,
					Msg:  "验证码错误",
				},
			}, nil
		}
	}

	user.GoogleEnabled = 0
	if err = l.svcCtx.UserModel.Update(l.ctx, user); err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}
