package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDefaultUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultUserBankLogic {
	return &SetDefaultUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置默认用户银行卡
func (l *SetDefaultUserBankLogic) SetDefaultUserBank(in *user.SetDefaultUserBankReq) (*user.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AdminCommonResp{}, nil
}
