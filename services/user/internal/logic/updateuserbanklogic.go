package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBankLogic {
	return &UpdateUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户银行卡
func (l *UpdateUserBankLogic) UpdateUserBank(in *user.UpdateUserBankReq) (*user.UpdateUserBankResp, error) {
	// todo: add your logic here and delete this line

	return &user.UpdateUserBankResp{}, nil
}
