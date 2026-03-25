package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserBankLogic {
	return &DeleteUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除用户银行卡
func (l *DeleteUserBankLogic) DeleteUserBank(in *user.DeleteUserBankReq) (*user.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AdminCommonResp{}, nil
}
