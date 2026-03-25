package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddUserBankLogic {
	return &AddUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 添加用户银行卡
func (l *AddUserBankLogic) AddUserBank(in *user.AddUserBankReq) (*user.AddUserBankResp, error) {
	// todo: add your logic here and delete this line

	return &user.AddUserBankResp{}, nil
}
