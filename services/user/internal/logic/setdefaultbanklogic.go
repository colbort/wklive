package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetDefaultBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetDefaultBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultBankLogic {
	return &SetDefaultBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置默认银行卡
func (l *SetDefaultBankLogic) SetDefaultBank(in *user.SetDefaultBankReq) (*user.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AppCommonResp{}, nil
}
