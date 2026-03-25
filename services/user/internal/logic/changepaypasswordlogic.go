package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePayPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChangePayPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePayPasswordLogic {
	return &ChangePayPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改支付密码
func (l *ChangePayPasswordLogic) ChangePayPassword(in *user.ChangePayPasswordReq) (*user.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AppCommonResp{}, nil
}
