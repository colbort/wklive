package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddBankLogic {
	return &AddBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 添加银行卡
func (l *AddBankLogic) AddBank(in *user.AddBankReq) (*user.AddBankResp, error) {
	// todo: add your logic here and delete this line

	return &user.AddBankResp{}, nil
}
