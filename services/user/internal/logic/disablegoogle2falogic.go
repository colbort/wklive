package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DisableGoogle2FALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDisableGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableGoogle2FALogic {
	return &DisableGoogle2FALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 禁用谷歌2FA
func (l *DisableGoogle2FALogic) DisableGoogle2FA(in *user.DisableGoogle2FAReq) (*user.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AppCommonResp{}, nil
}
