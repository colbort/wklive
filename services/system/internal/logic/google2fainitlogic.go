package logic

import (
	"context"

	"wklive/common/helper"
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
			Base: helper.GetErrResp(1, "用户不存在"),
		}, nil
	}

	secret, otpauthURL, _, err := utils.GenerateGoogle2FA(user.Username, utils.Default2FAIssuer, 256)
	if err != nil {
		return &system.Google2FAInitResp{
			Base: helper.GetErrResp(1, "生成2FA secret失败: "+err.Error()),
		}, err
	}

	// 将 secret 存储到 redis，设置过期时间，例如 10 分钟
	if err := l.svcCtx.UserModel.InsertGoogle2FASecret(l.ctx, in.UserId, secret); err != nil {
		return &system.Google2FAInitResp{
			Base: helper.GetErrResp(1, "存储2FA secret失败: "+err.Error()),
		}, err
	}
	return &system.Google2FAInitResp{
		Base:       helper.OkResp(),
		Secret:     secret,
		OtpauthUrl: otpauthURL,
	}, nil
}
