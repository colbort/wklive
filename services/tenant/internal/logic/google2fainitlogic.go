package logic

import (
	"context"

	"wklive/proto/tenant"
	"wklive/services/tenant/internal/svc"

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
func (l *Google2FAInitLogic) Google2FAInit(in *tenant.Google2FAInitReq) (*tenant.Google2FAInitResp, error) {
	// todo: add your logic here and delete this line

	return &tenant.Google2FAInitResp{}, nil
}
