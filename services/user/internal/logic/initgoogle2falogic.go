package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitGoogle2FALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitGoogle2FALogic {
	return &InitGoogle2FALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化谷歌2FA
func (l *InitGoogle2FALogic) InitGoogle2FA(in *user.InitGoogle2FAReq) (*user.InitGoogle2FAResp, error) {
	// todo: add your logic here and delete this line

	return &user.InitGoogle2FAResp{}, nil
}
