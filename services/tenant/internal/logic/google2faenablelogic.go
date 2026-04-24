package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type Google2FAEnableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGoogle2FAEnableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *Google2FAEnableLogic {
	return &Google2FAEnableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 启用Google 2FA
func (l *Google2FAEnableLogic) Google2FAEnable(in *tenant.Google2FAEnableReq) (*tenant.RespBase, error) {
	// todo: add your logic here and delete this line

	return &tenant.RespBase{}, nil
}
