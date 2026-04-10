package logic

import (
	"context"

	"wklive/common/helper"
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
	err := l.svcCtx.UserBankModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
