package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

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

// 禁用Google 2FA
func (l *Google2FADisableLogic) Google2FADisable(in *tenant.Google2FADisableReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
