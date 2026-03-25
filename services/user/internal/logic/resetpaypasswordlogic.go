package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPayPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPayPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPayPasswordLogic {
	return &ResetPayPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 重置支付密码
func (l *ResetPayPasswordLogic) ResetPayPassword(in *user.ResetPayPasswordReq) (*user.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AdminCommonResp{}, nil
}
