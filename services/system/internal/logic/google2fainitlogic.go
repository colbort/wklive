package logic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type Google2FAInitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGoogle2FAInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Google2FAInitLogic {
	return &Google2FAInitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 2FA
func (l *Google2FAInitLogic) Google2FAInit(in *system.Google2FAInitReq) (*system.Google2FAInitResp, error) {
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &system.Google2FAInitResp{
			Base: &system.RespBase{
				Code: 1,
				Msg:  "用户不存在",
			},
		}, nil
	}

	secret, otpauthURL, _, err := utils.GenerateGoogle2FA(user.Username, utils.Default2FAIssuer, 256)
	if err != nil {
		return &system.Google2FAInitResp{
			Base: &system.RespBase{
				Code: 1,
				Msg:  "生成2FA secret失败: " + err.Error(),
			},
		}, err
	}

	user.GoogleSecret = secret
	if err = l.svcCtx.UserModel.Update(l.ctx, user); err != nil {
		return nil, err
	}

	return &system.Google2FAInitResp{
		Base: &system.RespBase{
			Code: 200,
			Msg:  "初始化成功",
		},
		Secret:     secret,
		OtpauthUrl: otpauthURL,
	}, nil
}
