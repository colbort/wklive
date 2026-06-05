package logic

import (
	"context"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type TestVerificationCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewTestVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TestVerificationCodeLogic {
	return &TestVerificationCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 测试发送验证码
func (l *TestVerificationCodeLogic) TestVerificationCode(in *system.TestVerificationCodeReq) (*system.RespBase, error) {
	scene := in.Scene
	if scene == system.VerificationCodeScene_VERIFICATION_CODE_SCENE_UNKNOWN {
		scene = system.VerificationCodeScene_VERIFICATION_CODE_SCENE_TEST
	}
	sendLogic := NewSendVerificationCodeLogic(l.ctx, l.svcCtx)
	return sendLogic.SendVerificationCode(&system.SendVerificationCodeReq{
		TenantId: in.TenantId,
		Channel:  in.Channel,
		Email:    in.Email,
		Phone:    in.Phone,
		Scene:    scene,
	})
}
