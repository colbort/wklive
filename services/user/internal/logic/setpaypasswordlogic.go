package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetPayPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetPayPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetPayPasswordLogic {
	return &SetPayPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置支付密码
func (l *SetPayPasswordLogic) SetPayPassword(in *user.SetPayPasswordReq) (*user.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AppCommonResp{}, nil
}
