package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserBankLogic {
	return &GetUserBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户银行卡详情
func (l *GetUserBankLogic) GetUserBank(in *user.GetUserBankReq) (*user.GetUserBankResp, error) {
	// todo: add your logic here and delete this line

	return &user.GetUserBankResp{}, nil
}
