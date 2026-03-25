package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnableGoogle2FALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnableGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnableGoogle2FALogic {
	return &EnableGoogle2FALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 启用谷歌2FA
func (l *EnableGoogle2FALogic) EnableGoogle2FA(in *user.EnableGoogle2FAReq) (*user.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AppCommonResp{}, nil
}
