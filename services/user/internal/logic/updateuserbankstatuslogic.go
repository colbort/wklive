package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserBankStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserBankStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBankStatusLogic {
	return &UpdateUserBankStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新用户银行卡状态
func (l *UpdateUserBankStatusLogic) UpdateUserBankStatus(in *user.UpdateUserBankStatusReq) (*user.AdminCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AdminCommonResp{}, nil
}
