package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateBankLogic {
	return &UpdateBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新银行卡
func (l *UpdateBankLogic) UpdateBank(in *user.UpdateBankReq) (*user.UpdateBankResp, error) {
	// todo: add your logic here and delete this line

	return &user.UpdateBankResp{}, nil
}
