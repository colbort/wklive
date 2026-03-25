package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteBankLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteBankLogic {
	return &DeleteBankLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除银行卡
func (l *DeleteBankLogic) DeleteBank(in *user.DeleteBankReq) (*user.AppCommonResp, error) {
	// todo: add your logic here and delete this line

	return &user.AppCommonResp{}, nil
}
