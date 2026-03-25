package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetUserGoogle2FALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetUserGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetUserGoogle2FALogic {
	return &ResetUserGoogle2FALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 重置用户谷歌2FA
func (l *ResetUserGoogle2FALogic) ResetUserGoogle2FA(in *user.ResetUserGoogle2FAReq) (*user.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AdminCommonResp{}, nil
}
